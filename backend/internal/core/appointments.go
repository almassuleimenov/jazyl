package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/almasik/jazyl/internal/domain"
	"github.com/almasik/jazyl/internal/phone"
	"github.com/almasik/jazyl/internal/store"
)

// NewAppointment is input for creating an appointment.
type NewAppointment struct {
	PatientPhone string
	PatientName  string
	PatientLang  string
	DoctorID     *uuid.UUID
	ServiceID    *uuid.UUID
	ServiceName  string
	StartsAt     time.Time
	DurationMin  int
	Source       string
	LeadID       *uuid.UUID
	Notes        string
	Status       string // default scheduled
}

// CreateAppointment creates (or idempotently returns) an appointment and plans its reminders.
func (s *Service) CreateAppointment(ctx context.Context, clinicID uuid.UUID, in NewAppointment) (*domain.Appointment, bool, error) {
	var out *domain.Appointment
	var created bool
	err := s.store.InTx(ctx, func(tx *store.Store) error {
		c, err := tx.GetClinic(ctx, clinicID)
		if err != nil {
			return err
		}
		out, created, err = s.createAppointmentTx(ctx, tx, c, in)
		return err
	})
	if err != nil {
		return nil, false, err
	}
	if created {
		s.pub.Publish(clinicID, "appointment.created", out)
	}
	return out, created, nil
}

func (s *Service) createAppointmentTx(ctx context.Context, tx *store.Store, c *domain.Clinic, in NewAppointment) (*domain.Appointment, bool, error) {
	ph, err := phone.Normalize(in.PatientPhone)
	if err != nil {
		return nil, false, fmt.Errorf("%w: phone", domain.ErrInvalid)
	}
	if in.StartsAt.IsZero() {
		return nil, false, fmt.Errorf("%w: starts_at", domain.ErrInvalid)
	}
	if in.Source == "" {
		in.Source = "manual"
	}
	if in.Status == "" {
		in.Status = domain.ApptScheduled
	}
	if in.ServiceID != nil {
		svc, err := tx.GetService(ctx, c.ID, *in.ServiceID)
		if err != nil {
			return nil, false, fmt.Errorf("%w: service_id", domain.ErrInvalid)
		}
		if in.ServiceName == "" {
			in.ServiceName = svc.Name
		}
		if in.DurationMin == 0 {
			in.DurationMin = svc.DurationMin
		}
	}
	if in.DurationMin == 0 {
		in.DurationMin = c.Settings.SlotMinutes
	}

	p, err := tx.UpsertPatient(ctx, c.ID, ph, strings.TrimSpace(in.PatientName), in.PatientLang)
	if err != nil {
		return nil, false, err
	}
	a, created, err := tx.CreateAppointment(ctx, &domain.Appointment{
		ClinicID: c.ID, PatientID: p.ID, DoctorID: in.DoctorID, ServiceID: in.ServiceID,
		ServiceName: strings.TrimSpace(in.ServiceName), StartsAt: in.StartsAt, DurationMin: in.DurationMin,
		Status: in.Status, Source: in.Source, LeadID: in.LeadID, Notes: in.Notes,
	})
	if err != nil {
		return nil, false, err
	}
	if created {
		if err := tx.RecordAppointmentEvent(ctx, c.ID, a.ID, "", a.Status, "staff"); err != nil {
			return nil, false, err
		}
		if err := s.planAppointmentReminders(ctx, tx, c, a); err != nil {
			return nil, false, err
		}
	}
	return a, created, nil
}

// ChangeStatus moves an appointment to a new status and adjusts reminders.
// actor: "staff", "patient" or "system".
func (s *Service) ChangeStatus(ctx context.Context, clinicID, apptID uuid.UUID, status, actor string) (*domain.Appointment, error) {
	if !domain.ValidApptStatus(status) {
		return nil, fmt.Errorf("%w: status", domain.ErrInvalid)
	}
	var out *domain.Appointment
	err := s.store.InTx(ctx, func(tx *store.Store) error {
		c, err := tx.GetClinic(ctx, clinicID)
		if err != nil {
			return err
		}
		a, err := s.changeStatusTx(ctx, tx, c, apptID, status, actor)
		out = a
		return err
	})
	if err != nil {
		return nil, err
	}
	s.pub.Publish(clinicID, "appointment.updated", out)
	return out, nil
}

func (s *Service) changeStatusTx(ctx context.Context, tx *store.Store, c *domain.Clinic, apptID uuid.UUID, status, actor string) (*domain.Appointment, error) {
	if err := tx.SetAppointmentStatus(ctx, c.ID, apptID, status, actor); err != nil {
		return nil, err
	}
	a, err := tx.GetAppointment(ctx, c.ID, apptID)
	if err != nil {
		return nil, err
	}
	switch status {
	case domain.ApptCancelled, domain.ApptNoShow, domain.ApptRescheduleRequested:
		if err := tx.SkipPendingReminders(ctx, a.ID, "status "+status); err != nil {
			return nil, err
		}
		if err := tx.CancelQueuedForAppointment(ctx, a.ID); err != nil {
			return nil, err
		}
	case domain.ApptConfirmed:
		if err := tx.SkipPendingReminders(ctx, a.ID, "confirmed", domain.ReminderDayBefore); err != nil {
			return nil, err
		}
	case domain.ApptCompleted:
		if err := s.planReview(ctx, tx, c, a); err != nil {
			return nil, err
		}
	case domain.ApptScheduled:
		if err := s.planAppointmentReminders(ctx, tx, c, a); err != nil {
			return nil, err
		}
	}
	return a, nil
}

func (s *Service) planReview(ctx context.Context, tx *store.Store, c *domain.Clinic, a *domain.Appointment) error {
	if !c.Settings.ReviewRequests {
		return nil
	}
	return tx.PlanReminder(ctx, c.ID, a.ID, domain.ReminderReview, ReviewTime(a.EndsAt(), c.Settings, c.Location()))
}

// AppointmentPatch edits an appointment. Nil fields are unchanged.
type AppointmentPatch struct {
	StartsAt    *time.Time
	DurationMin *int
	DoctorID    *uuid.UUID
	ClearDoctor bool
	ServiceID   *uuid.UUID
	ServiceName *string
	Notes       *string
}

// UpdateAppointment edits an appointment; moving it in time resets it to "scheduled" and replans reminders.
func (s *Service) UpdateAppointment(ctx context.Context, clinicID, apptID uuid.UUID, p AppointmentPatch) (*domain.Appointment, error) {
	var out *domain.Appointment
	err := s.store.InTx(ctx, func(tx *store.Store) error {
		c, err := tx.GetClinic(ctx, clinicID)
		if err != nil {
			return err
		}
		a, err := tx.LockAppointment(ctx, clinicID, apptID)
		if err != nil {
			return err
		}
		moved := p.StartsAt != nil && !p.StartsAt.Equal(a.StartsAt)
		if p.StartsAt != nil {
			a.StartsAt = *p.StartsAt
		}
		if p.DurationMin != nil {
			a.DurationMin = *p.DurationMin
		}
		if p.ClearDoctor {
			a.DoctorID = nil
		} else if p.DoctorID != nil {
			a.DoctorID = p.DoctorID
		}
		if p.ServiceID != nil {
			a.ServiceID = p.ServiceID
		}
		if p.ServiceName != nil {
			a.ServiceName = *p.ServiceName
		}
		if p.Notes != nil {
			a.Notes = *p.Notes
		}
		if err := tx.UpdateAppointment(ctx, a); err != nil {
			return err
		}
		if moved {
			if a.Status != domain.ApptScheduled {
				if err := tx.SetAppointmentStatus(ctx, clinicID, a.ID, domain.ApptScheduled, "staff"); err != nil {
					return err
				}
				a.Status = domain.ApptScheduled
			}
			if err := tx.CancelQueuedForAppointment(ctx, a.ID); err != nil {
				return err
			}
			if err := s.planAppointmentReminders(ctx, tx, c, a); err != nil {
				return err
			}
		}
		out, err = tx.GetAppointment(ctx, clinicID, apptID)
		return err
	})
	if err != nil {
		return nil, err
	}
	s.pub.Publish(clinicID, "appointment.updated", out)
	return out, nil
}

// AutoComplete marks finished confirmed visits as completed and schedules review requests.
func (s *Service) AutoComplete(ctx context.Context) (int, error) {
	var n int
	err := s.store.InTx(ctx, func(tx *store.Store) error {
		done, err := tx.AutoCompleteConfirmed(ctx, s.now().Add(-time.Hour))
		if err != nil {
			return err
		}
		clinics := map[uuid.UUID]*domain.Clinic{}
		for _, a := range done {
			c, ok := clinics[a.ClinicID]
			if !ok {
				if c, err = tx.GetClinic(ctx, a.ClinicID); err != nil {
					return err
				}
				clinics[a.ClinicID] = c
			}
			if err := s.planReview(ctx, tx, c, a); err != nil {
				return err
			}
		}
		n = len(done)
		return nil
	})
	return n, err
}
