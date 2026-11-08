package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Metrics are the numbers a clinic owner sees in reports and the dashboard.
type Metrics struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`

	Appointments         int            `json:"appointments"`
	AppointmentsBySource map[string]int `json:"appointments_by_source"`
	RemindersSent        int            `json:"reminders_sent"`
	Confirmed            int            `json:"confirmed"`
	RescheduleRequested  int            `json:"reschedule_requested"`
	CancelledByPatient   int            `json:"cancelled_by_patient"`
	Completed            int            `json:"completed"`
	NoShows              int            `json:"no_shows"`
	// PreventedNoShows: the patient warned the clinic through the bot instead of silently not coming.
	PreventedNoShows int     `json:"prevented_no_shows"`
	SavedRevenueKZT  int     `json:"saved_revenue_kzt"`
	ConfirmationRate float64 `json:"confirmation_rate"` // confirmed / reminder replies requested
	NoShowRate       float64 `json:"no_show_rate"`

	Leads         int            `json:"leads"`
	LeadsBySource map[string]int `json:"leads_by_source"`
	LeadsBooked   int            `json:"leads_booked"`

	Reviews       int     `json:"reviews"`
	AvgRating     float64 `json:"avg_rating"`
	ReviewsTo2GIS int     `json:"reviews_to_2gis"`
	Complaints    int     `json:"complaints"`

	BotChats         int `json:"bot_chats"` // unique patient phones that wrote to the bot
	InboundMessages  int `json:"inbound_messages"`
	OutboundMessages int `json:"outbound_messages"`
	FailedMessages   int `json:"failed_messages"`
}

// ClinicMetrics computes metrics for [from, to).
func (s *Store) ClinicMetrics(ctx context.Context, clinicID uuid.UUID, from, to time.Time, avgCheck int) (*Metrics, error) {
	m := &Metrics{From: from, To: to, AppointmentsBySource: map[string]int{}, LeadsBySource: map[string]int{}}

	err := s.db.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM appointments WHERE clinic_id=$1 AND starts_at >= $2 AND starts_at < $3 AND status <> 'cancelled'),
			(SELECT count(*) FROM messages WHERE clinic_id=$1 AND direction='out' AND kind='reminder' AND status='sent' AND sent_at >= $2 AND sent_at < $3),
			(SELECT count(*) FROM appointment_events WHERE clinic_id=$1 AND to_status='confirmed' AND created_at >= $2 AND created_at < $3),
			(SELECT count(*) FROM appointment_events WHERE clinic_id=$1 AND to_status='reschedule_requested' AND actor='patient' AND created_at >= $2 AND created_at < $3),
			(SELECT count(*) FROM appointment_events WHERE clinic_id=$1 AND to_status='cancelled' AND actor='patient' AND created_at >= $2 AND created_at < $3),
			(SELECT count(*) FROM appointment_events WHERE clinic_id=$1 AND to_status='completed' AND created_at >= $2 AND created_at < $3),
			(SELECT count(*) FROM appointment_events WHERE clinic_id=$1 AND to_status='no_show' AND created_at >= $2 AND created_at < $3),
			(SELECT count(*) FROM leads WHERE clinic_id=$1 AND created_at >= $2 AND created_at < $3),
			(SELECT count(*) FROM leads WHERE clinic_id=$1 AND created_at >= $2 AND created_at < $3 AND status='booked'),
			(SELECT count(*) FROM reviews WHERE clinic_id=$1 AND created_at >= $2 AND created_at < $3),
			(SELECT COALESCE(avg(rating), 0)::float8 FROM reviews WHERE clinic_id=$1 AND created_at >= $2 AND created_at < $3),
			(SELECT count(*) FROM reviews WHERE clinic_id=$1 AND created_at >= $2 AND created_at < $3 AND routed_to='2gis'),
			(SELECT count(*) FROM reviews WHERE clinic_id=$1 AND created_at >= $2 AND created_at < $3 AND routed_to='owner'),
			(SELECT count(DISTINCT phone) FROM messages WHERE clinic_id=$1 AND direction='in' AND kind='patient_in' AND created_at >= $2 AND created_at < $3),
			(SELECT count(*) FROM messages WHERE clinic_id=$1 AND direction='in' AND created_at >= $2 AND created_at < $3),
			(SELECT count(*) FROM messages WHERE clinic_id=$1 AND direction='out' AND status='sent' AND sent_at >= $2 AND sent_at < $3),
			(SELECT count(*) FROM messages WHERE clinic_id=$1 AND direction='out' AND status='failed' AND created_at >= $2 AND created_at < $3)`,
		clinicID, from, to).Scan(
		&m.Appointments, &m.RemindersSent, &m.Confirmed, &m.RescheduleRequested, &m.CancelledByPatient, &m.Completed, &m.NoShows,
		&m.Leads, &m.LeadsBooked, &m.Reviews, &m.AvgRating, &m.ReviewsTo2GIS, &m.Complaints,
		&m.BotChats, &m.InboundMessages, &m.OutboundMessages, &m.FailedMessages)
	if err != nil {
		return nil, mapErr(err)
	}

	if err := s.groupCount(ctx, m.AppointmentsBySource, `SELECT source, count(*) FROM appointments
		WHERE clinic_id=$1 AND starts_at >= $2 AND starts_at < $3 AND status <> 'cancelled' GROUP BY source`, clinicID, from, to); err != nil {
		return nil, err
	}
	if err := s.groupCount(ctx, m.LeadsBySource, `SELECT source, count(*) FROM leads
		WHERE clinic_id=$1 AND created_at >= $2 AND created_at < $3 GROUP BY source`, clinicID, from, to); err != nil {
		return nil, err
	}

	m.PreventedNoShows = m.RescheduleRequested + m.CancelledByPatient
	m.SavedRevenueKZT = m.PreventedNoShows * avgCheck
	if answered := m.Confirmed + m.PreventedNoShows + m.NoShows; answered > 0 {
		m.ConfirmationRate = round2(float64(m.Confirmed) / float64(answered))
	}
	if finished := m.Completed + m.NoShows; finished > 0 {
		m.NoShowRate = round2(float64(m.NoShows) / float64(finished))
	}
	m.AvgRating = round2(m.AvgRating)
	return m, nil
}

func (s *Store) groupCount(ctx context.Context, dst map[string]int, q string, args ...any) error {
	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return mapErr(err)
	}
	defer rows.Close()
	for rows.Next() {
		var k string
		var n int
		if err := rows.Scan(&k, &n); err != nil {
			return mapErr(err)
		}
		dst[k] = n
	}
	return mapErr(rows.Err())
}

func round2(f float64) float64 { return float64(int(f*100+0.5)) / 100 }

// DailyPoint is one day of the dashboard timeseries.
type DailyPoint struct {
	Day           string `json:"day"` // YYYY-MM-DD in the clinic time zone
	Appointments  int    `json:"appointments"`
	RemindersSent int    `json:"reminders_sent"`
	Confirmed     int    `json:"confirmed"`
	Prevented     int    `json:"prevented"`
	Leads         int    `json:"leads"`
	Inbound       int    `json:"inbound"`
}

// DailySeries returns per-day counters in the clinic's time zone.
func (s *Store) DailySeries(ctx context.Context, clinicID uuid.UUID, from, to time.Time, tz string) ([]DailyPoint, error) {
	rows, err := s.db.Query(ctx, `
		WITH days AS (
			SELECT generate_series(date_trunc('day', $2::timestamptz AT TIME ZONE $4::text),
				date_trunc('day', ($3::timestamptz - interval '1 microsecond') AT TIME ZONE $4::text), interval '1 day')::date AS d
		)
		SELECT to_char(days.d, 'YYYY-MM-DD'),
			(SELECT count(*) FROM appointments a WHERE a.clinic_id=$1 AND a.status <> 'cancelled' AND (a.starts_at AT TIME ZONE $4)::date = days.d),
			(SELECT count(*) FROM messages m WHERE m.clinic_id=$1 AND m.direction='out' AND m.kind='reminder' AND m.status='sent' AND (m.sent_at AT TIME ZONE $4)::date = days.d),
			(SELECT count(*) FROM appointment_events e WHERE e.clinic_id=$1 AND e.to_status='confirmed' AND (e.created_at AT TIME ZONE $4)::date = days.d),
			(SELECT count(*) FROM appointment_events e WHERE e.clinic_id=$1 AND e.actor='patient' AND e.to_status IN ('reschedule_requested','cancelled') AND (e.created_at AT TIME ZONE $4)::date = days.d),
			(SELECT count(*) FROM leads l WHERE l.clinic_id=$1 AND (l.created_at AT TIME ZONE $4)::date = days.d),
			(SELECT count(*) FROM messages m WHERE m.clinic_id=$1 AND m.direction='in' AND (m.created_at AT TIME ZONE $4)::date = days.d)
		FROM days ORDER BY days.d`, clinicID, from, to, tz)
	if err != nil {
		return nil, mapErr(err)
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (DailyPoint, error) {
		var p DailyPoint
		err := r.Scan(&p.Day, &p.Appointments, &p.RemindersSent, &p.Confirmed, &p.Prevented, &p.Leads, &p.Inbound)
		return p, err
	})
}

// ChatSummary is one row of "who talks to the bot".
type ChatSummary struct {
	Phone         string     `json:"phone"`
	PatientID     *uuid.UUID `json:"patient_id"`
	PatientName   string     `json:"patient_name"`
	InCount       int        `json:"in_count"`
	OutCount      int        `json:"out_count"`
	LastMessageAt time.Time  `json:"last_message_at"`
	LastBody      string     `json:"last_body"`
	LastDirection string     `json:"last_direction"`
}

// ListChats aggregates conversations by phone for [from, to).
func (s *Store) ListChats(ctx context.Context, clinicID uuid.UUID, from, to time.Time, limit int) ([]ChatSummary, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.Query(ctx, `
		WITH agg AS (
			SELECT phone,
				count(*) FILTER (WHERE direction='in') AS in_count,
				count(*) FILTER (WHERE direction='out') AS out_count,
				max(created_at) AS last_at
			FROM messages
			WHERE clinic_id=$1 AND created_at >= $2 AND created_at < $3 AND kind NOT IN ('admin_in','admin_notice','owner_notice','report')
			GROUP BY phone
		)
		SELECT agg.phone, p.id, COALESCE(p.name, ''), agg.in_count, agg.out_count, agg.last_at, last.body, last.direction
		FROM agg
		LEFT JOIN patients p ON p.clinic_id=$1 AND p.phone=agg.phone
		CROSS JOIN LATERAL (
			SELECT body, direction FROM messages m WHERE m.clinic_id=$1 AND m.phone=agg.phone ORDER BY created_at DESC LIMIT 1
		) last
		ORDER BY agg.last_at DESC LIMIT $4`, clinicID, from, to, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (ChatSummary, error) {
		var c ChatSummary
		err := r.Scan(&c.Phone, &c.PatientID, &c.PatientName, &c.InCount, &c.OutCount, &c.LastMessageAt, &c.LastBody, &c.LastDirection)
		return c, err
	})
}

// PlatformClinicRow is a superadmin overview line.
type PlatformClinicRow struct {
	ClinicID        uuid.UUID  `json:"clinic_id"`
	Name            string     `json:"name"`
	City            string     `json:"city"`
	Plan            string     `json:"plan"`
	Status          string     `json:"status"`
	WASessionStatus string     `json:"wa_session_status"`
	Appointments30d int        `json:"appointments_30d"`
	Reminders30d    int        `json:"reminders_30d"`
	Failed30d       int        `json:"failed_30d"`
	LastInboundAt   *time.Time `json:"last_inbound_at"`
}

// PlatformOverview lists every clinic with 30-day activity counters.
func (s *Store) PlatformOverview(ctx context.Context) ([]PlatformClinicRow, error) {
	rows, err := s.db.Query(ctx, `
		SELECT c.id, c.name, c.city, c.plan, c.status, c.wa_session_status,
			(SELECT count(*) FROM appointments a WHERE a.clinic_id=c.id AND a.created_at > now() - interval '30 days'),
			(SELECT count(*) FROM messages m WHERE m.clinic_id=c.id AND m.kind='reminder' AND m.status='sent' AND m.sent_at > now() - interval '30 days'),
			(SELECT count(*) FROM messages m WHERE m.clinic_id=c.id AND m.status='failed' AND m.created_at > now() - interval '30 days'),
			(SELECT max(created_at) FROM messages m WHERE m.clinic_id=c.id AND m.direction='in')
		FROM clinics c ORDER BY c.created_at DESC`)
	if err != nil {
		return nil, mapErr(err)
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (PlatformClinicRow, error) {
		var v PlatformClinicRow
		err := r.Scan(&v.ClinicID, &v.Name, &v.City, &v.Plan, &v.Status, &v.WASessionStatus, &v.Appointments30d, &v.Reminders30d, &v.Failed30d, &v.LastInboundAt)
		return v, err
	})
}
