package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/almasik/jazyl/internal/domain"
	"github.com/almasik/jazyl/internal/messaging"
	"github.com/almasik/jazyl/internal/observability"
	"github.com/almasik/jazyl/internal/store"
)

// ReminderPlan is when each reminder of an appointment should go out (nil = not planned).
type ReminderPlan struct {
	DayBefore *time.Time
	TwoHours  *time.Time
}

// InQuiet reports whether t falls into the clinic's quiet hours (no messages to patients).
func InQuiet(t time.Time, s domain.Settings, loc *time.Location) bool {
	h := t.In(loc).Hour()
	if s.QuietStartHour > s.QuietEndHour {
		return h >= s.QuietStartHour || h < s.QuietEndHour
	}
	return h >= s.QuietStartHour && h < s.QuietEndHour
}

// nextAwake moves t forward to the end of quiet hours if needed.
func nextAwake(t time.Time, s domain.Settings, loc *time.Location) time.Time {
	if !InQuiet(t, s, loc) {
		return t
	}
	l := t.In(loc)
	wake := time.Date(l.Year(), l.Month(), l.Day(), s.QuietEndHour, 0, 0, 0, loc)
	if !wake.After(l) {
		wake = wake.AddDate(0, 0, 1)
	}
	return wake
}

// PlanReminders decides reminder times for an appointment starting at starts, given now.
//
//   - day_before: the evening before at ReminderEveningHour (default 20:00). If that moment has
//     passed (booked late), it is sent as soon as possible outside quiet hours, but only if it
//     still lands at least 2.5h before the visit.
//   - two_hours: 2h before the visit, unless that falls in quiet hours or is within 3h after
//     the day_before message (avoid two pings in a row).
func PlanReminders(starts, now time.Time, s domain.Settings, loc *time.Location) ReminderPlan {
	var p ReminderPlan
	ls := starts.In(loc)
	prev := ls.AddDate(0, 0, -1)
	evening := time.Date(prev.Year(), prev.Month(), prev.Day(), s.ReminderEveningHour, 0, 0, 0, loc)

	if evening.After(now) {
		e := evening
		p.DayBefore = &e
	} else {
		asap := nextAwake(now.Add(time.Minute), s, loc)
		if asap.Before(starts.Add(-150 * time.Minute)) {
			p.DayBefore = &asap
		}
	}

	if s.TwoHourReminder {
		t := starts.Add(-2 * time.Hour)
		ok := t.After(now) && !InQuiet(t, s, loc)
		if ok && p.DayBefore != nil && t.Sub(*p.DayBefore) < 3*time.Hour {
			ok = false
		}
		if ok {
			p.TwoHours = &t
		}
	}
	return p
}

// ReviewTime is when to ask for a rating: 2h after the visit ends, outside quiet hours.
func ReviewTime(end time.Time, s domain.Settings, loc *time.Location) time.Time {
	return nextAwake(end.Add(2*time.Hour), s, loc)
}

// planAppointmentReminders (re)plans reminders for an active appointment.
func (s *Service) planAppointmentReminders(ctx context.Context, tx *store.Store, c *domain.Clinic, a *domain.Appointment) error {
	plan := PlanReminders(a.StartsAt, s.now(), c.Settings, c.Location())
	if plan.DayBefore != nil && a.Status == domain.ApptScheduled {
		if err := tx.PlanReminder(ctx, c.ID, a.ID, domain.ReminderDayBefore, *plan.DayBefore); err != nil {
			return err
		}
	} else if err := tx.SkipPendingReminders(ctx, a.ID, "not applicable", domain.ReminderDayBefore); err != nil {
		return err
	}
	if plan.TwoHours != nil {
		if err := tx.PlanReminder(ctx, c.ID, a.ID, domain.ReminderTwoHours, *plan.TwoHours); err != nil {
			return err
		}
	} else if err := tx.SkipPendingReminders(ctx, a.ID, "not applicable", domain.ReminderTwoHours); err != nil {
		return err
	}
	return nil
}

// ProcessDueReminders turns due reminders into outbox messages. Returns how many were handled.
func (s *Service) ProcessDueReminders(ctx context.Context, limit int) (int, error) {
	handled := 0
	err := s.store.InTx(ctx, func(tx *store.Store) error {
		due, err := tx.ClaimDueReminders(ctx, limit)
		if err != nil {
			return err
		}
		for _, r := range due {
			msgID, reason, err := s.dispatchReminder(ctx, tx, r)
			if err != nil {
				return fmt.Errorf("reminder %s: %w", r.ID, err)
			}
			if err := tx.FinishReminder(ctx, r.ID, msgID, reason); err != nil {
				return err
			}
			outcome := "sent"
			if msgID == nil {
				outcome = "skipped"
			}
			observability.RemindersProcessed.WithLabelValues(r.Kind, outcome).Inc()
			handled++
		}
		return nil
	})
	return handled, err
}

// dispatchReminder returns the queued message id, or nil and a skip reason.
func (s *Service) dispatchReminder(ctx context.Context, tx *store.Store, r store.DueReminder) (*uuid.UUID, string, error) {
	c, err := tx.GetClinic(ctx, r.ClinicID)
	if err != nil {
		return nil, "", err
	}
	if c.Status != "pilot" && c.Status != "active" {
		return nil, "clinic " + c.Status, nil
	}
	a, err := tx.LockAppointment(ctx, r.ClinicID, r.AppointmentID)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, "appointment deleted", nil
	}
	if err != nil {
		return nil, "", err
	}
	p, err := tx.GetPatient(ctx, c.ID, a.PatientID)
	if err != nil {
		return nil, "", err
	}
	if p.OptedOutAt != nil {
		return nil, "patient opted out", nil
	}
	now := s.now()
	v := visitFor(c, p, a)
	apptID := a.ID

	switch r.Kind {
	case domain.ReminderDayBefore, domain.ReminderTwoHours:
		if a.Status != domain.ApptScheduled && a.Status != domain.ApptConfirmed {
			return nil, "status " + a.Status, nil
		}
		if !a.StartsAt.After(now.Add(30 * time.Minute)) {
			return nil, "too late", nil
		}
		var body string
		askReply := true
		switch {
		case r.Kind == domain.ReminderTwoHours:
			body = messaging.TwoHours(v, a.Status == domain.ApptConfirmed)
			askReply = a.Status != domain.ApptConfirmed
		case a.Status == domain.ApptConfirmed:
			return nil, "already confirmed", nil
		case isTomorrow(now, a.StartsAt, c):
			body = messaging.DayBefore(v)
		default:
			body = messaging.SameDay(v)
		}
		m, err := s.toPatient(ctx, tx, c, p, &apptID, domain.MsgReminder, body)
		if err != nil {
			return nil, "", err
		}
		if askReply {
			ttl := a.EndsAt().Sub(now) + time.Hour
			if err := tx.SetConversation(ctx, c.ID, p.Phone, domain.AwaitReminderReply, &apptID, ttl); err != nil {
				return nil, "", err
			}
		}
		return &m.ID, "", nil

	case domain.ReminderReview:
		if a.Status != domain.ApptCompleted {
			return nil, "status " + a.Status, nil
		}
		if !c.Settings.ReviewRequests {
			return nil, "reviews disabled", nil
		}
		m, err := s.toPatient(ctx, tx, c, p, &apptID, domain.MsgReviewRequest, messaging.ReviewRequest(v))
		if err != nil {
			return nil, "", err
		}
		if err := tx.SetConversation(ctx, c.ID, p.Phone, domain.AwaitReviewRating, &apptID, 48*time.Hour); err != nil {
			return nil, "", err
		}
		return &m.ID, "", nil
	}
	return nil, "unknown kind", nil
}

func isTomorrow(now, starts time.Time, c *domain.Clinic) bool {
	loc := c.Location()
	n, st := now.In(loc), starts.In(loc)
	tomorrow := time.Date(n.Year(), n.Month(), n.Day()+1, 0, 0, 0, 0, loc)
	return st.Year() == tomorrow.Year() && st.YearDay() == tomorrow.YearDay()
}
