package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/almasik/jazyl/internal/domain"
	"github.com/almasik/jazyl/internal/messaging"
	"github.com/almasik/jazyl/internal/observability"
	"github.com/almasik/jazyl/internal/phone"
	"github.com/almasik/jazyl/internal/store"
)

// ErrPlanRequired is returned when the clinic's tariff does not include the feature.
var ErrPlanRequired = errors.New("feature not included in plan")

// SiteBooking is a booking request coming from a clinic's website.
type SiteBooking struct {
	Name        string
	Phone       string
	ServiceID   *uuid.UUID
	Service     string
	PreferredAt *time.Time
	Comment     string
	PageURL     string
	UTM         map[string]string
	Source      string // site | instagram | 2gis ... (default site)
	Language    string
}

// BookingResult tells the website what happened.
type BookingResult struct {
	LeadID        uuid.UUID  `json:"lead_id"`
	Status        string     `json:"status"` // "received" | "booked"
	AppointmentID *uuid.UUID `json:"appointment_id,omitempty"`
	StartsAt      *time.Time `json:"starts_at,omitempty"`
}

// CreateSiteBooking stores a lead, notifies the administrator and (if enabled and the slot is free)
// books the appointment straight away.
func (s *Service) CreateSiteBooking(ctx context.Context, clinicID uuid.UUID, in SiteBooking) (*BookingResult, error) {
	ph, err := phone.Normalize(in.Phone)
	if err != nil {
		return nil, fmt.Errorf("%w: phone", domain.ErrInvalid)
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Source == "" {
		in.Source = "site"
	}
	if in.PreferredAt != nil && in.PreferredAt.Before(s.now().Add(-time.Hour)) {
		return nil, fmt.Errorf("%w: preferred_at is in the past", domain.ErrInvalid)
	}

	var res BookingResult
	var lead *domain.Lead
	err = s.store.InTx(ctx, func(tx *store.Store) error {
		c, err := tx.GetClinic(ctx, clinicID)
		if err != nil {
			return err
		}
		if !domain.PlanAllowsSite(c.Plan) && c.Status != "pilot" {
			return ErrPlanRequired
		}
		if in.ServiceID != nil {
			svc, err := tx.GetService(ctx, c.ID, *in.ServiceID)
			if err != nil {
				return fmt.Errorf("%w: service_id", domain.ErrInvalid)
			}
			in.Service = svc.Name
		}

		lead, err = tx.CreateLead(ctx, &domain.Lead{
			ClinicID: c.ID, Source: in.Source, Name: in.Name, Phone: ph, Service: in.Service,
			PreferredAt: in.PreferredAt, Comment: in.Comment, PageURL: in.PageURL, UTM: in.UTM,
		})
		if err != nil {
			return err
		}
		res = BookingResult{LeadID: lead.ID, Status: "received"}

		// The patient agreed to be contacted by submitting the form (the site shows the consent text).
		p, err := tx.UpsertPatient(ctx, c.ID, ph, in.Name, in.Language)
		if err != nil {
			return err
		}
		if err := tx.SetPatientConsent(ctx, c.ID, p.ID); err != nil {
			return err
		}

		if c.Settings.AutoBookFromSite && in.PreferredAt != nil {
			dur := c.Settings.SlotMinutes
			if in.ServiceID != nil {
				if svc, err := tx.GetService(ctx, c.ID, *in.ServiceID); err == nil {
					dur = svc.DurationMin
				}
			}
			free, err := s.slotFree(ctx, tx, c, *in.PreferredAt, dur)
			if err != nil {
				return err
			}
			if free {
				leadID := lead.ID
				a, _, err := s.createAppointmentTx(ctx, tx, c, NewAppointment{
					PatientPhone: ph, PatientName: in.Name, PatientLang: in.Language, ServiceID: in.ServiceID,
					ServiceName: in.Service, StartsAt: *in.PreferredAt, DurationMin: dur, Source: "site", LeadID: &leadID,
				})
				if err != nil {
					return err
				}
				if err := tx.UpdateLeadStatus(ctx, c.ID, lead.ID, "booked", &a.ID); err != nil {
					return err
				}
				if _, err := s.toPatient(ctx, tx, c, p, &a.ID, domain.MsgAutoReply, messaging.BookingConfirmed(visitFor(c, p, a))); err != nil {
					return err
				}
				if err := s.toAdmins(ctx, tx, c, messaging.AdminAutoBooked(in.Name, ph, a.StartsAt.In(c.Location()))); err != nil {
					return err
				}
				res.Status, res.AppointmentID, res.StartsAt = "booked", &a.ID, &a.StartsAt
				return nil
			}
		}

		var pref *time.Time
		if in.PreferredAt != nil {
			t := in.PreferredAt.In(c.Location())
			pref = &t
		}
		return s.toAdmins(ctx, tx, c, messaging.AdminNewLead(in.Source, in.Name, ph, in.Service, pref, in.Comment))
	})
	if err != nil {
		return nil, err
	}
	observability.Leads.WithLabelValues(in.Source).Inc()
	s.pub.Publish(clinicID, "lead.created", lead)
	return &res, nil
}

// ConvertLead books an appointment for a lead (from the dashboard).
func (s *Service) ConvertLead(ctx context.Context, clinicID, leadID uuid.UUID, startsAt time.Time, doctorID, serviceID *uuid.UUID) (*domain.Appointment, error) {
	var out *domain.Appointment
	err := s.store.InTx(ctx, func(tx *store.Store) error {
		c, err := tx.GetClinic(ctx, clinicID)
		if err != nil {
			return err
		}
		l, err := tx.GetLead(ctx, clinicID, leadID)
		if err != nil {
			return err
		}
		if l.Status == "booked" && l.AppointmentID != nil {
			return fmt.Errorf("%w: lead already booked", domain.ErrConflict)
		}
		lid := l.ID
		a, _, err := s.createAppointmentTx(ctx, tx, c, NewAppointment{
			PatientPhone: l.Phone, PatientName: l.Name, DoctorID: doctorID, ServiceID: serviceID,
			ServiceName: l.Service, StartsAt: startsAt, Source: "site", LeadID: &lid,
		})
		if err != nil {
			return err
		}
		out = a
		return tx.UpdateLeadStatus(ctx, clinicID, leadID, "booked", &a.ID)
	})
	if err != nil {
		return nil, err
	}
	s.pub.Publish(clinicID, "appointment.created", out)
	return out, nil
}

var weekdayKeys = [...]string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}

// Slot is a free time slot.
type Slot struct {
	StartsAt time.Time `json:"starts_at"`
	Time     string    `json:"time"` // HH:MM local
}

// FreeSlots lists free slots on a local date for a service duration.
func (s *Service) FreeSlots(ctx context.Context, clinicID uuid.UUID, date time.Time, serviceID *uuid.UUID) ([]Slot, error) {
	c, err := s.store.GetClinic(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	dur := c.Settings.SlotMinutes
	if serviceID != nil {
		svc, err := s.store.GetService(ctx, clinicID, *serviceID)
		if err != nil {
			return nil, fmt.Errorf("%w: service_id", domain.ErrInvalid)
		}
		dur = svc.DurationMin
	}
	loc := c.Location()
	day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
	open, closeT, ok := workingWindow(c.Settings, day)
	if !ok {
		return []Slot{}, nil
	}
	busy, err := s.store.BusyIntervals(ctx, clinicID, open, closeT)
	if err != nil {
		return nil, err
	}
	return ComputeSlots(open, closeT, s.now().Add(time.Hour), time.Duration(c.Settings.SlotMinutes)*time.Minute,
		time.Duration(dur)*time.Minute, c.Settings.Chairs, busy, loc), nil
}

func workingWindow(st domain.Settings, day time.Time) (time.Time, time.Time, bool) {
	h, ok := st.WorkingHours[weekdayKeys[day.Weekday()]]
	if !ok {
		return time.Time{}, time.Time{}, false
	}
	o, err1 := time.Parse("15:04", h.Open)
	cl, err2 := time.Parse("15:04", h.Close)
	if err1 != nil || err2 != nil {
		return time.Time{}, time.Time{}, false
	}
	open := time.Date(day.Year(), day.Month(), day.Day(), o.Hour(), o.Minute(), 0, 0, day.Location())
	closeT := time.Date(day.Year(), day.Month(), day.Day(), cl.Hour(), cl.Minute(), 0, 0, day.Location())
	return open, closeT, closeT.After(open)
}

// ComputeSlots is the pure slot algorithm: a slot is free if fewer than chairs busy intervals overlap it.
func ComputeSlots(open, closeT, notBefore time.Time, step, dur time.Duration, chairs int, busy [][2]time.Time, loc *time.Location) []Slot {
	out := []Slot{}
	for t := open; !t.Add(dur).After(closeT); t = t.Add(step) {
		if t.Before(notBefore) {
			continue
		}
		end := t.Add(dur)
		overlaps := 0
		for _, b := range busy {
			if b[0].Before(end) && b[1].After(t) {
				overlaps++
			}
		}
		if overlaps < chairs {
			out = append(out, Slot{StartsAt: t, Time: t.In(loc).Format("15:04")})
		}
	}
	return out
}

func (s *Service) slotFree(ctx context.Context, tx *store.Store, c *domain.Clinic, at time.Time, durMin int) (bool, error) {
	loc := c.Location()
	l := at.In(loc)
	open, closeT, ok := workingWindow(c.Settings, time.Date(l.Year(), l.Month(), l.Day(), 0, 0, 0, 0, loc))
	end := at.Add(time.Duration(durMin) * time.Minute)
	if !ok || at.Before(open) || end.After(closeT) || at.Before(s.now()) {
		return false, nil
	}
	busy, err := tx.BusyIntervals(ctx, c.ID, at, end)
	if err != nil {
		return false, err
	}
	return len(busy) < c.Settings.Chairs, nil
}
