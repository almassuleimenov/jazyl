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

const msgCols = `id, clinic_id, patient_id, appointment_id, direction, phone, kind, body, status, provider_message_id,
	attempts, last_error, created_at, sent_at`

func scanMsg(row pgx.Row) (*domain.Message, error) {
	var m domain.Message
	if err := row.Scan(&m.ID, &m.ClinicID, &m.PatientID, &m.AppointmentID, &m.Direction, &m.Phone, &m.Kind, &m.Body, &m.Status,
		&m.ProviderMessageID, &m.Attempts, &m.LastError, &m.CreatedAt, &m.SentAt); err != nil {
		return nil, mapErr(err)
	}
	return &m, nil
}

// OutMessage describes a message to enqueue.
type OutMessage struct {
	ClinicID      uuid.UUID
	PatientID     *uuid.UUID
	AppointmentID *uuid.UUID
	Phone         string
	Kind          string
	Body          string
	NotBefore     time.Time
}

// EnqueueMessage inserts an outbound message into the outbox.
func (s *Store) EnqueueMessage(ctx context.Context, m OutMessage) (*domain.Message, error) {
	nb := m.NotBefore
	if nb.IsZero() {
		nb = time.Now()
	}
	return scanMsg(s.db.QueryRow(ctx, `INSERT INTO messages (clinic_id, patient_id, appointment_id, direction, phone, kind, body, status, next_attempt_at)
		VALUES ($1,$2,$3,'out',$4,$5,$6,'queued',$7) RETURNING `+msgCols,
		m.ClinicID, m.PatientID, m.AppointmentID, m.Phone, m.Kind, m.Body, nb))
}

// InsertInbound logs a received message.
func (s *Store) InsertInbound(ctx context.Context, clinicID uuid.UUID, patientID *uuid.UUID, phone, kind, body, providerID string) (*domain.Message, error) {
	return scanMsg(s.db.QueryRow(ctx, `INSERT INTO messages (clinic_id, patient_id, direction, phone, kind, body, status, provider_message_id, sent_at)
		VALUES ($1,$2,'in',$3,$4,$5,'received',NULLIF($6,''),now()) RETURNING `+msgCols,
		clinicID, patientID, phone, kind, body, providerID))
}

// OutboxItem is a claimed outbound message with the routing data the sender needs.
type OutboxItem struct {
	Message      *domain.Message
	Provider     string
	SessionID    *string
	ClinicName   string
	ClinicStatus string
}

// ClaimOutbox leases up to limit due messages for lease. Rows are skipped if locked by another worker.
func (s *Store) ClaimOutbox(ctx context.Context, limit int, lease time.Duration) ([]OutboxItem, error) {
	rows, err := s.db.Query(ctx, `
		WITH due AS (
			SELECT m.id FROM messages m
			WHERE m.status = 'queued' AND m.next_attempt_at <= now()
			ORDER BY m.next_attempt_at
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE messages m SET status='sending', locked_until = now() + make_interval(secs => $2), attempts = attempts + 1
		FROM due WHERE m.id = due.id
		RETURNING m.id, m.clinic_id, m.patient_id, m.appointment_id, m.direction, m.phone, m.kind, m.body, m.status,
			m.provider_message_id, m.attempts, m.last_error, m.created_at, m.sent_at`, limit, lease.Seconds())
	if err != nil {
		return nil, mapErr(err)
	}
	var msgs []*domain.Message
	for rows.Next() {
		m, err := scanMsg(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		msgs = append(msgs, m)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, mapErr(err)
	}

	out := make([]OutboxItem, 0, len(msgs))
	for _, m := range msgs {
		it := OutboxItem{Message: m}
		if err := s.db.QueryRow(ctx, `SELECT wa_provider, wa_session_id, name, status FROM clinics WHERE id=$1`, m.ClinicID).
			Scan(&it.Provider, &it.SessionID, &it.ClinicName, &it.ClinicStatus); err != nil {
			return nil, mapErr(err)
		}
		out = append(out, it)
	}
	return out, nil
}

// MarkMessageSent finalises a sent message.
func (s *Store) MarkMessageSent(ctx context.Context, id uuid.UUID, providerID string) error {
	_, err := s.db.Exec(ctx, `UPDATE messages SET status='sent', provider_message_id=NULLIF($2,''), sent_at=now(), locked_until=NULL, last_error=''
		WHERE id=$1`, id, providerID)
	return mapErr(err)
}

// RetryMessage puts a message back into the queue at a later time.
func (s *Store) RetryMessage(ctx context.Context, id uuid.UUID, at time.Time, reason string, refundAttempt bool) error {
	_, err := s.db.Exec(ctx, `UPDATE messages SET status='queued', next_attempt_at=$2, last_error=$3, locked_until=NULL,
		attempts = CASE WHEN $4 THEN GREATEST(attempts - 1, 0) ELSE attempts END WHERE id=$1`, id, at, truncate(reason, 500), refundAttempt)
	return mapErr(err)
}

// FailMessage marks a message as permanently failed.
func (s *Store) FailMessage(ctx context.Context, id uuid.UUID, reason string) error {
	_, err := s.db.Exec(ctx, `UPDATE messages SET status='failed', last_error=$2, locked_until=NULL WHERE id=$1`, id, truncate(reason, 500))
	return mapErr(err)
}

// ReleaseStuckMessages re-queues messages whose lease expired (worker crash).
func (s *Store) ReleaseStuckMessages(ctx context.Context) (int64, error) {
	tag, err := s.db.Exec(ctx, `UPDATE messages SET status='queued', locked_until=NULL WHERE status='sending' AND locked_until < now()`)
	return tag.RowsAffected(), mapErr(err)
}

// CancelQueuedForAppointment cancels unsent messages tied to an appointment (e.g. it was cancelled).
func (s *Store) CancelQueuedForAppointment(ctx context.Context, apptID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `UPDATE messages SET status='cancelled' WHERE appointment_id=$1 AND status='queued'`, apptID)
	return mapErr(err)
}

// ListMessages lists a clinic's messages, optionally for one phone.
func (s *Store) ListMessages(ctx context.Context, clinicID uuid.UUID, phone string, page Page) ([]*domain.Message, error) {
	ts, id, err := page.keysetArgs()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, `SELECT `+msgCols+` FROM messages
		WHERE clinic_id=$1 AND ($2 = '' OR phone=$2) AND (created_at, id) < ($3, $4)
		ORDER BY created_at DESC, id DESC LIMIT $5`, clinicID, phone, ts, id, page.limit())
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	var out []*domain.Message
	for rows.Next() {
		m, err := scanMsg(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, mapErr(rows.Err())
}

// ---------- reminders ----------

// PlanReminder creates or reschedules a pending reminder. Already-sent reminders are left alone.
func (s *Store) PlanReminder(ctx context.Context, clinicID, apptID uuid.UUID, kind string, at time.Time) error {
	_, err := s.db.Exec(ctx, `INSERT INTO reminders (clinic_id, appointment_id, kind, send_at) VALUES ($1,$2,$3,$4)
		ON CONFLICT (appointment_id, kind) DO UPDATE SET send_at=EXCLUDED.send_at, status='pending', skip_reason=''
		WHERE reminders.status <> 'sent'`, clinicID, apptID, kind, at)
	return mapErr(err)
}

// SkipPendingReminders skips every pending reminder of an appointment (optionally only some kinds).
func (s *Store) SkipPendingReminders(ctx context.Context, apptID uuid.UUID, reason string, kinds ...string) error {
	if kinds == nil {
		kinds = []string{} // a nil slice would be sent as NULL and match nothing
	}
	_, err := s.db.Exec(ctx, `UPDATE reminders SET status='skipped', skip_reason=$2, processed_at=now()
		WHERE appointment_id=$1 AND status='pending' AND (cardinality($3::text[]) = 0 OR kind = ANY($3))`, apptID, reason, kinds)
	return mapErr(err)
}

// DueReminder is a reminder ready to be processed.
type DueReminder struct {
	ID            uuid.UUID
	ClinicID      uuid.UUID
	AppointmentID uuid.UUID
	Kind          string
	SendAt        time.Time
}

// ClaimDueReminders locks due reminders (run inside InTx; they are finalised in the same tx).
func (s *Store) ClaimDueReminders(ctx context.Context, limit int) ([]DueReminder, error) {
	rows, err := s.db.Query(ctx, `SELECT id, clinic_id, appointment_id, kind, send_at FROM reminders
		WHERE status='pending' AND send_at <= now() ORDER BY send_at LIMIT $1 FOR UPDATE SKIP LOCKED`, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (DueReminder, error) {
		var d DueReminder
		err := r.Scan(&d.ID, &d.ClinicID, &d.AppointmentID, &d.Kind, &d.SendAt)
		return d, err
	})
}

// FinishReminder marks a reminder as sent (with its message) or skipped.
func (s *Store) FinishReminder(ctx context.Context, id uuid.UUID, msgID *uuid.UUID, skipReason string) error {
	status := "sent"
	if msgID == nil {
		status = "skipped"
	}
	_, err := s.db.Exec(ctx, `UPDATE reminders SET status=$2, message_id=$3, skip_reason=$4, processed_at=now() WHERE id=$1`,
		id, status, msgID, skipReason)
	return mapErr(err)
}

// ---------- conversations ----------

// Conversation is what the bot awaits from a phone.
type Conversation struct {
	Awaiting      string
	AppointmentID *uuid.UUID
}

// SetConversation stores the awaited reply for a phone.
func (s *Store) SetConversation(ctx context.Context, clinicID uuid.UUID, phone, awaiting string, apptID *uuid.UUID, ttl time.Duration) error {
	_, err := s.db.Exec(ctx, `INSERT INTO conversations (clinic_id, phone, awaiting, appointment_id, expires_at)
		VALUES ($1,$2,$3,$4, now() + make_interval(secs => $5))
		ON CONFLICT (clinic_id, phone) DO UPDATE SET awaiting=EXCLUDED.awaiting, appointment_id=EXCLUDED.appointment_id,
			expires_at=EXCLUDED.expires_at, updated_at=now()`, clinicID, phone, awaiting, apptID, ttl.Seconds())
	return mapErr(err)
}

// GetConversation returns the non-expired conversation state, or ErrNotFound.
func (s *Store) GetConversation(ctx context.Context, clinicID uuid.UUID, phone string) (*Conversation, error) {
	var c Conversation
	err := s.db.QueryRow(ctx, `SELECT awaiting, appointment_id FROM conversations WHERE clinic_id=$1 AND phone=$2 AND expires_at > now()`,
		clinicID, phone).Scan(&c.Awaiting, &c.AppointmentID)
	if err != nil {
		return nil, mapErr(err)
	}
	return &c, nil
}

// ClearConversation removes conversation state.
func (s *Store) ClearConversation(ctx context.Context, clinicID uuid.UUID, phone string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM conversations WHERE clinic_id=$1 AND phone=$2`, clinicID, phone)
	return mapErr(err)
}

// ---------- reviews ----------

// CreateReview stores a rating (one per appointment; a repeat updates it).
func (s *Store) CreateReview(ctx context.Context, r *domain.Review) (*domain.Review, error) {
	var out domain.Review
	err := s.db.QueryRow(ctx, `INSERT INTO reviews (clinic_id, appointment_id, patient_id, rating, comment, routed_to)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (appointment_id) DO UPDATE SET rating=EXCLUDED.rating, routed_to=EXCLUDED.routed_to
		RETURNING id, clinic_id, appointment_id, patient_id, rating, comment, routed_to, created_at`,
		r.ClinicID, r.AppointmentID, r.PatientID, r.Rating, r.Comment, r.RoutedTo).
		Scan(&out.ID, &out.ClinicID, &out.AppointmentID, &out.PatientID, &out.Rating, &out.Comment, &out.RoutedTo, &out.CreatedAt)
	return &out, mapErr(err)
}

// SetReviewComment attaches the complaint text to an appointment's review.
func (s *Store) SetReviewComment(ctx context.Context, clinicID, apptID uuid.UUID, comment string) error {
	_, err := s.db.Exec(ctx, `UPDATE reviews SET comment=$3 WHERE clinic_id=$1 AND appointment_id=$2`, clinicID, apptID, truncate(comment, 2000))
	return mapErr(err)
}

// ListReviews lists reviews newest first.
func (s *Store) ListReviews(ctx context.Context, clinicID uuid.UUID, page Page) ([]*domain.Review, error) {
	ts, id, err := page.keysetArgs()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, `SELECT id, clinic_id, appointment_id, patient_id, rating, comment, routed_to, created_at FROM reviews
		WHERE clinic_id=$1 AND (created_at, id) < ($2, $3) ORDER BY created_at DESC, id DESC LIMIT $4`, clinicID, ts, id, page.limit())
	if err != nil {
		return nil, mapErr(err)
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (*domain.Review, error) {
		var v domain.Review
		err := r.Scan(&v.ID, &v.ClinicID, &v.AppointmentID, &v.PatientID, &v.Rating, &v.Comment, &v.RoutedTo, &v.CreatedAt)
		return &v, err
	})
}

// ---------- schedule imports ----------

// InsertScheduleImport audits a notebook import.
func (s *Store) InsertScheduleImport(ctx context.Context, clinicID uuid.UUID, fromPhone, raw string, date time.Time, parsed int, failures any) error {
	b, err := json.Marshal(failures)
	if err != nil {
		return fmt.Errorf("encode failures: %w", err)
	}
	_, err = s.db.Exec(ctx, `INSERT INTO schedule_imports (clinic_id, from_phone, raw_text, target_date, parsed_count, failed_lines)
		VALUES ($1,$2,$3,$4,$5,$6)`, clinicID, fromPhone, truncate(raw, 20000), date, parsed, b)
	return mapErr(err)
}

// ---------- reports ----------

// ReportRow is a stored report.
type ReportRow struct {
	ID          uuid.UUID       `json:"id"`
	ClinicID    uuid.UUID       `json:"clinic_id"`
	Period      string          `json:"period"`
	PeriodStart time.Time       `json:"period_start"`
	PeriodEnd   time.Time       `json:"period_end"`
	Metrics     json.RawMessage `json:"metrics"`
	MessageID   *uuid.UUID      `json:"message_id"`
	CreatedAt   time.Time       `json:"created_at"`
}

// InsertReport stores a report once per (clinic, period, start). Returns ErrConflict if it already exists.
func (s *Store) InsertReport(ctx context.Context, clinicID uuid.UUID, period string, start, end time.Time, metrics any, msgID *uuid.UUID) error {
	b, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("encode metrics: %w", err)
	}
	tag, err := s.db.Exec(ctx, `INSERT INTO reports (clinic_id, period, period_start, period_end, metrics, message_id)
		VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT (clinic_id, period, period_start) DO NOTHING`, clinicID, period, start, end, b, msgID)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConflict
	}
	return nil
}

// ReportExists reports whether a report was already produced.
func (s *Store) ReportExists(ctx context.Context, clinicID uuid.UUID, period string, start time.Time) (bool, error) {
	var ok bool
	err := s.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM reports WHERE clinic_id=$1 AND period=$2 AND period_start=$3)`,
		clinicID, period, start).Scan(&ok)
	return ok, mapErr(err)
}

// ListReports lists a clinic's reports.
func (s *Store) ListReports(ctx context.Context, clinicID uuid.UUID, limit int) ([]ReportRow, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.db.Query(ctx, `SELECT id, clinic_id, period, period_start, period_end, metrics, message_id, created_at FROM reports
		WHERE clinic_id=$1 ORDER BY period_start DESC, period LIMIT $2`, clinicID, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (ReportRow, error) {
		var v ReportRow
		err := r.Scan(&v.ID, &v.ClinicID, &v.Period, &v.PeriodStart, &v.PeriodEnd, &v.Metrics, &v.MessageID, &v.CreatedAt)
		return v, err
	})
}

// ---------- webhook dedupe ----------

// MarkWebhookSeen returns true the first time a key is seen.
func (s *Store) MarkWebhookSeen(ctx context.Context, key string) (bool, error) {
	tag, err := s.db.Exec(ctx, `INSERT INTO webhook_events (idempotency_key) VALUES ($1) ON CONFLICT DO NOTHING`, truncate(key, 300))
	if err != nil {
		return false, mapErr(err)
	}
	return tag.RowsAffected() == 1, nil
}

// Housekeeping deletes stale rows (webhook keys, expired tokens and conversations).
func (s *Store) Housekeeping(ctx context.Context) error {
	var errs []error
	for _, q := range []string{
		`DELETE FROM webhook_events WHERE received_at < now() - interval '7 days'`,
		`DELETE FROM refresh_tokens WHERE expires_at < now() - interval '7 days'`,
		`DELETE FROM conversations WHERE expires_at < now()`,
	} {
		if _, err := s.db.Exec(ctx, q); err != nil {
			errs = append(errs, mapErr(err))
		}
	}
	return errors.Join(errs...)
}
