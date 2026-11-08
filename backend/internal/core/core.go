// Package core contains Jazyl's business logic: appointments and reminders, the WhatsApp bot
// (patient replies, reviews, notebook-mode imports), website bookings and owner reports.
package core

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/almasik/jazyl/internal/domain"
	"github.com/almasik/jazyl/internal/messaging"
	"github.com/almasik/jazyl/internal/realtime"
	"github.com/almasik/jazyl/internal/store"
)

// Service is the business-logic facade used by HTTP handlers and background jobs.
type Service struct {
	store *store.Store
	pub   realtime.Publisher
	log   *slog.Logger
	now   func() time.Time
}

// New creates the service.
func New(st *store.Store, pub realtime.Publisher, log *slog.Logger) *Service {
	if pub == nil {
		pub = realtime.Nop{}
	}
	return &Service{store: st, pub: pub, log: log.With("component", "core"), now: time.Now}
}

// Store exposes the store for read-only handlers.
func (s *Service) Store() *store.Store { return s.store }

// ---------- notifications (enqueue into the outbox, in the caller's transaction) ----------

func (s *Service) toPatient(ctx context.Context, tx *store.Store, c *domain.Clinic, p *domain.Patient, apptID *uuid.UUID, kind, body string) (*domain.Message, error) {
	if c.WASessionID == nil && c.WAProvider != "mock" {
		s.log.Warn("clinic has no whatsapp session; message queued until it is linked", "clinic_id", c.ID)
	}
	pid := p.ID
	return tx.EnqueueMessage(ctx, store.OutMessage{
		ClinicID: c.ID, PatientID: &pid, AppointmentID: apptID, Phone: p.Phone, Kind: kind, Body: body,
	})
}

func (s *Service) toAdmins(ctx context.Context, tx *store.Store, c *domain.Clinic, body string) error {
	targets := c.AdminPhones
	if len(targets) == 0 && c.OwnerPhone != "" {
		targets = []string{c.OwnerPhone}
	}
	for _, ph := range targets {
		if _, err := tx.EnqueueMessage(ctx, store.OutMessage{ClinicID: c.ID, Phone: ph, Kind: domain.MsgAdminNotice, Body: body}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) toOwner(ctx context.Context, tx *store.Store, c *domain.Clinic, kind, body string) (*domain.Message, error) {
	if c.OwnerPhone == "" {
		return nil, nil
	}
	return tx.EnqueueMessage(ctx, store.OutMessage{ClinicID: c.ID, Phone: c.OwnerPhone, Kind: kind, Body: body})
}

func visitFor(c *domain.Clinic, p *domain.Patient, a *domain.Appointment) messaging.Visit {
	lang := p.Language
	if lang == "" {
		lang = c.Language
	}
	return messaging.Visit{
		Lang:        lang,
		PatientName: p.Name,
		ClinicName:  c.Name,
		Address:     c.Address,
		StartsAt:    a.StartsAt.In(c.Location()),
	}
}
