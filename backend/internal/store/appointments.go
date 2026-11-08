package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/almasik/jazyl/internal/domain"
)

const apptCols = `a.id, a.clinic_id, a.patient_id, a.doctor_id, a.service_id, a.service_name, a.starts_at, a.duration_min,
	a.status, a.source, a.lead_id, a.notes, a.confirmed_at, a.status_changed_at, a.created_at, a.updated_at,
	p.name, p.phone, COALESCE(d.name, '')`

const apptFrom = ` FROM appointments a JOIN patients p ON p.id = a.patient_id LEFT JOIN doctors d ON d.id = a.doctor_id `

func scanAppt(row pgx.Row) (*domain.Appointment, error) {
	var a domain.Appointment
	err := row.Scan(&a.ID, &a.ClinicID, &a.PatientID, &a.DoctorID, &a.ServiceID, &a.ServiceName, &a.StartsAt, &a.DurationMin,
		&a.Status, &a.Source, &a.LeadID, &a.Notes, &a.ConfirmedAt, &a.StatusChangedAt, &a.CreatedAt, &a.UpdatedAt,
		&a.PatientName, &a.PatientPhone, &a.DoctorName)
	if err != nil {
		return nil, mapErr(err)
	}
	return &a, nil
}

// CreateAppointment inserts an appointment. If the same patient is already booked at the same
// minute (not cancelled), the existing row is returned with created=false (idempotent imports).
func (s *Store) CreateAppointment(ctx context.Context, a *domain.Appointment) (appt *domain.Appointment, created bool, err error) {
	var id uuid.UUID
	err = s.db.QueryRow(ctx, `
		INSERT INTO appointments (clinic_id, patient_id, doctor_id, service_id, service_name, starts_at, duration_min, status, source, lead_id, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (clinic_id, patient_id, starts_at) WHERE status NOT IN ('cancelled') DO NOTHING
		RETURNING id`,
		a.ClinicID, a.PatientID, a.DoctorID, a.ServiceID, a.ServiceName, a.StartsAt, a.DurationMin, a.Status, a.Source, a.LeadID, a.Notes).
		Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		existing, err := scanAppt(s.db.QueryRow(ctx, `SELECT `+apptCols+apptFrom+`
			WHERE a.clinic_id=$1 AND a.patient_id=$2 AND a.starts_at=$3 AND a.status <> 'cancelled'`, a.ClinicID, a.PatientID, a.StartsAt))
		return existing, false, err
	}
	if err != nil {
		return nil, false, mapErr(err)
	}
	appt, err = s.GetAppointment(ctx, a.ClinicID, id)
	return appt, true, err
}

// GetAppointment returns one appointment with patient/doctor names.
func (s *Store) GetAppointment(ctx context.Context, clinicID, id uuid.UUID) (*domain.Appointment, error) {
	return scanAppt(s.db.QueryRow(ctx, `SELECT `+apptCols+apptFrom+` WHERE a.clinic_id=$1 AND a.id=$2`, clinicID, id))
}

// LockAppointment selects an appointment FOR UPDATE (must run inside InTx).
func (s *Store) LockAppointment(ctx context.Context, clinicID, id uuid.UUID) (*domain.Appointment, error) {
	return scanAppt(s.db.QueryRow(ctx, `SELECT `+apptCols+apptFrom+` WHERE a.clinic_id=$1 AND a.id=$2 FOR UPDATE OF a`, clinicID, id))
}

// ApptFilter filters the appointment list.
type ApptFilter struct {
	From, To  time.Time
	Status    string
	DoctorID  *uuid.UUID
	PatientID *uuid.UUID
}

// ListAppointments lists appointments in a time range ordered by start.
func (s *Store) ListAppointments(ctx context.Context, clinicID uuid.UUID, f ApptFilter) ([]*domain.Appointment, error) {
	rows, err := s.db.Query(ctx, `SELECT `+apptCols+apptFrom+`
		WHERE a.clinic_id=$1 AND a.starts_at >= $2 AND a.starts_at < $3
		  AND ($4 = '' OR a.status = $4)
		  AND ($5::uuid IS NULL OR a.doctor_id = $5)
		  AND ($6::uuid IS NULL OR a.patient_id = $6)
		ORDER BY a.starts_at LIMIT 1000`, clinicID, f.From, f.To, f.Status, f.DoctorID, f.PatientID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	var out []*domain.Appointment
	for rows.Next() {
		a, err := scanAppt(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, mapErr(rows.Err())
}

// SetAppointmentStatus changes the status and records the transition in appointment_events.
// actor is "patient", "staff" or "system".
func (s *Store) SetAppointmentStatus(ctx context.Context, clinicID, id uuid.UUID, status, actor string) error {
	tag, err := s.db.Exec(ctx, `
		WITH prev AS (
			SELECT id, status FROM appointments WHERE clinic_id=$1 AND id=$2 FOR UPDATE
		), upd AS (
			UPDATE appointments a SET status=$3, status_changed_at=now(), updated_at=now(),
				confirmed_at = CASE WHEN $3 = 'confirmed' THEN COALESCE(a.confirmed_at, now()) ELSE a.confirmed_at END
			FROM prev WHERE a.id = prev.id AND prev.status <> $3
			RETURNING a.id
		)
		INSERT INTO appointment_events (clinic_id, appointment_id, from_status, to_status, actor)
		SELECT $1, prev.id, prev.status, $3, $4 FROM prev JOIN upd ON upd.id = prev.id`, clinicID, id, status, actor)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		// Either missing or already in that status; distinguish for callers.
		var exists bool
		if err := s.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM appointments WHERE clinic_id=$1 AND id=$2)`, clinicID, id).Scan(&exists); err != nil {
			return mapErr(err)
		}
		if !exists {
			return domain.ErrNotFound
		}
	}
	return nil
}

// RecordAppointmentEvent records a transition that happened outside SetAppointmentStatus (e.g. bulk updates).
func (s *Store) RecordAppointmentEvent(ctx context.Context, clinicID, apptID uuid.UUID, from, to, actor string) error {
	_, err := s.db.Exec(ctx, `INSERT INTO appointment_events (clinic_id, appointment_id, from_status, to_status, actor) VALUES ($1,$2,$3,$4,$5)`,
		clinicID, apptID, from, to, actor)
	return mapErr(err)
}

// UpdateAppointment edits time, doctor, service and notes.
func (s *Store) UpdateAppointment(ctx context.Context, a *domain.Appointment) error {
	tag, err := s.db.Exec(ctx, `UPDATE appointments SET starts_at=$3, duration_min=$4, doctor_id=$5, service_id=$6,
		service_name=$7, notes=$8, updated_at=now() WHERE clinic_id=$1 AND id=$2`,
		a.ClinicID, a.ID, a.StartsAt, a.DurationMin, a.DoctorID, a.ServiceID, a.ServiceName, a.Notes)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// BusyIntervals returns [start,end) of active appointments overlapping a window (slot calculation).
func (s *Store) BusyIntervals(ctx context.Context, clinicID uuid.UUID, from, to time.Time) ([][2]time.Time, error) {
	rows, err := s.db.Query(ctx, `SELECT starts_at, starts_at + make_interval(mins => duration_min) FROM appointments
		WHERE clinic_id=$1 AND status IN ('scheduled','confirmed','reschedule_requested')
		  AND starts_at < $3 AND starts_at + make_interval(mins => duration_min) > $2`, clinicID, from, to)
	if err != nil {
		return nil, mapErr(err)
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) ([2]time.Time, error) {
		var iv [2]time.Time
		err := r.Scan(&iv[0], &iv[1])
		return iv, err
	})
}

// AutoCompleteConfirmed marks confirmed appointments that ended before cutoff as completed.
// Only clinics with auto_complete_confirmed enabled are affected. Returns the affected rows.
func (s *Store) AutoCompleteConfirmed(ctx context.Context, cutoff time.Time) ([]*domain.Appointment, error) {
	rows, err := s.db.Query(ctx, `
		WITH done AS (
			UPDATE appointments a SET status='completed', status_changed_at=now(), updated_at=now()
			FROM clinics c
			WHERE c.id = a.clinic_id AND a.status = 'confirmed'
			  AND COALESCE((c.settings->>'auto_complete_confirmed')::boolean, true)
			  AND a.starts_at + make_interval(mins => a.duration_min) < $1
			  AND a.starts_at > $1 - interval '3 days'
			RETURNING a.id, a.clinic_id
		), ev AS (
			INSERT INTO appointment_events (clinic_id, appointment_id, from_status, to_status, actor)
			SELECT clinic_id, id, 'confirmed', 'completed', 'system' FROM done
		)
		SELECT `+apptCols+apptFrom+` JOIN done ON done.id = a.id`, cutoff)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	var out []*domain.Appointment
	for rows.Next() {
		a, err := scanAppt(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, mapErr(rows.Err())
}

// ---------- leads ----------

const leadCols = `id, clinic_id, source, name, phone, service, preferred_at, comment, status, page_url, utm, appointment_id, created_at, updated_at`

func scanLead(row pgx.Row) (*domain.Lead, error) {
	var l domain.Lead
	var utm []byte
	if err := row.Scan(&l.ID, &l.ClinicID, &l.Source, &l.Name, &l.Phone, &l.Service, &l.PreferredAt, &l.Comment, &l.Status,
		&l.PageURL, &utm, &l.AppointmentID, &l.CreatedAt, &l.UpdatedAt); err != nil {
		return nil, mapErr(err)
	}
	l.UTM = map[string]string{}
	if len(utm) > 0 {
		if err := json.Unmarshal(utm, &l.UTM); err != nil {
			return nil, fmt.Errorf("decode utm: %w", err)
		}
	}
	return &l, nil
}

// CreateLead inserts a lead.
func (s *Store) CreateLead(ctx context.Context, l *domain.Lead) (*domain.Lead, error) {
	if l.UTM == nil {
		l.UTM = map[string]string{}
	}
	utm, err := json.Marshal(l.UTM)
	if err != nil {
		return nil, fmt.Errorf("encode utm: %w", err)
	}
	if l.Status == "" {
		l.Status = "new"
	}
	return scanLead(s.db.QueryRow(ctx, `INSERT INTO leads (clinic_id, source, name, phone, service, preferred_at, comment, status, page_url, utm)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING `+leadCols,
		l.ClinicID, l.Source, l.Name, l.Phone, l.Service, l.PreferredAt, l.Comment, l.Status, l.PageURL, utm))
}

// GetLead returns a lead.
func (s *Store) GetLead(ctx context.Context, clinicID, id uuid.UUID) (*domain.Lead, error) {
	return scanLead(s.db.QueryRow(ctx, `SELECT `+leadCols+` FROM leads WHERE clinic_id=$1 AND id=$2`, clinicID, id))
}

// ListLeads lists leads, newest first.
func (s *Store) ListLeads(ctx context.Context, clinicID uuid.UUID, status string, page Page) ([]*domain.Lead, error) {
	ts, id, err := page.keysetArgs()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, `SELECT `+leadCols+` FROM leads
		WHERE clinic_id=$1 AND ($2 = '' OR status=$2) AND (created_at, id) < ($3, $4)
		ORDER BY created_at DESC, id DESC LIMIT $5`, clinicID, status, ts, id, page.limit())
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	var out []*domain.Lead
	for rows.Next() {
		l, err := scanLead(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, mapErr(rows.Err())
}

// UpdateLeadStatus sets status and optionally links an appointment.
func (s *Store) UpdateLeadStatus(ctx context.Context, clinicID, id uuid.UUID, status string, apptID *uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `UPDATE leads SET status=$3, appointment_id=COALESCE($4, appointment_id), updated_at=now()
		WHERE clinic_id=$1 AND id=$2`, clinicID, id, status, apptID)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
