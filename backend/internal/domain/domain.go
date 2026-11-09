// Package domain defines Jazyl's core entities, statuses and domain errors.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Domain errors mapped to HTTP statuses by the API layer.
var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrInvalid      = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

// Plans (tariffs).
const (
	PlanNoShow    = "no_show"    // «Без неявок» 15 000 ₸
	PlanFullChair = "full_chair" // «Полное кресло» 29 000 ₸
	PlanGrowth    = "growth"     // «Рост» 49 000 ₸
)

// PlanPriceKZT returns the monthly price of a plan.
func PlanPriceKZT(plan string) int {
	switch plan {
	case PlanFullChair:
		return 29000
	case PlanGrowth:
		return 49000
	default:
		return 15000
	}
}

// PlanAllowsSite reports whether a plan includes the website booking integration.
func PlanAllowsSite(plan string) bool { return plan == PlanFullChair || plan == PlanGrowth }

// Roles.
const (
	RoleSuperadmin = "superadmin" // Jazyl team, sees every clinic
	RoleOwner      = "owner"      // clinic owner
	RoleAdmin      = "admin"      // clinic administrator
)

// Appointment statuses.
const (
	ApptScheduled           = "scheduled"
	ApptConfirmed           = "confirmed"
	ApptRescheduleRequested = "reschedule_requested"
	ApptCancelled           = "cancelled"
	ApptCompleted           = "completed"
	ApptNoShow              = "no_show"
)

// ValidApptStatus reports whether s is a known appointment status.
func ValidApptStatus(s string) bool {
	switch s {
	case ApptScheduled, ApptConfirmed, ApptRescheduleRequested, ApptCancelled, ApptCompleted, ApptNoShow:
		return true
	}
	return false
}

// Reminder kinds.
const (
	ReminderDayBefore = "day_before"
	ReminderTwoHours  = "two_hours"
	ReminderReview    = "review_request"
)

// Message kinds (used for analytics breakdowns).
const (
	MsgReminder      = "reminder"
	MsgReviewRequest = "review_request"
	MsgReviewLink    = "review_link"
	MsgAdminNotice   = "admin_notice"
	MsgOwnerNotice   = "owner_notice"
	MsgReport        = "report"
	MsgAutoReply     = "auto_reply"
	MsgPatientIn     = "patient_in"
	MsgAdminIn       = "admin_in"
	MsgManual        = "manual"
)

// Conversation "awaiting" states.
const (
	AwaitReminderReply = "reminder_reply"
	AwaitReviewRating  = "review_rating"
	AwaitReviewComment = "review_comment"
)

// Clinic is a tenant.
type Clinic struct {
	ID              uuid.UUID `json:"id"`
	Slug            string    `json:"slug"`
	Name            string    `json:"name"`
	City            string    `json:"city"`
	Address         string    `json:"address"`
	Timezone        string    `json:"timezone"`
	Language        string    `json:"language"`
	Plan            string    `json:"plan"`
	Status          string    `json:"status"`
	OwnerPhone      string    `json:"owner_phone"`
	AdminPhones     []string  `json:"admin_phones"`
	GisReviewURL    string    `json:"gis_review_url"`
	AvgCheckKZT     int       `json:"avg_check_kzt"`
	Settings        Settings  `json:"settings"`
	WAProvider      string    `json:"wa_provider"`
	WASessionID     *string   `json:"wa_session_id"`
	WASessionStatus string    `json:"wa_session_status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Location returns the clinic's time zone (falls back to Asia/Almaty, UTC+5).
func (c *Clinic) Location() *time.Location {
	if loc, err := time.LoadLocation(c.Timezone); err == nil {
		return loc
	}
	return time.FixedZone("UTC+5", 5*3600)
}

// IsAdminPhone reports whether phone belongs to the clinic staff (admins or owner).
// Operating reports whether the clinic's automation and staff access are on.
// "paused" and "churned" switch off reminders, the bot, outgoing messages and staff logins;
// the data stays, and setting the status back to pilot/active restores everything.
func (c *Clinic) Operating() bool { return ClinicOperating(c.Status) }

// ClinicOperating is Operating for a bare status value.
func ClinicOperating(status string) bool { return status == "pilot" || status == "active" }

func (c *Clinic) IsAdminPhone(phone string) bool {
	if phone == "" {
		return false
	}
	if phone == c.OwnerPhone {
		return true
	}
	for _, p := range c.AdminPhones {
		if p == phone {
			return true
		}
	}
	return false
}

// DayHours is an opening interval in "HH:MM" local time.
type DayHours struct {
	Open  string `json:"open"`
	Close string `json:"close"`
}

// Settings are per-clinic behaviour knobs stored as JSONB.
type Settings struct {
	ReminderEveningHour   int                 `json:"reminder_evening_hour"`
	TwoHourReminder       bool                `json:"two_hour_reminder"`
	QuietStartHour        int                 `json:"quiet_start_hour"`
	QuietEndHour          int                 `json:"quiet_end_hour"`
	Chairs                int                 `json:"chairs"`
	SlotMinutes           int                 `json:"slot_minutes"`
	WorkingHours          map[string]DayHours `json:"working_hours"` // keys: mon..sun
	AutoBookFromSite      bool                `json:"auto_book_from_site"`
	AutoCompleteConfirmed bool                `json:"auto_complete_confirmed"`
	ReviewRequests        bool                `json:"review_requests"`
	WeeklyReport          bool                `json:"weekly_report"`
	MonthlyReport         bool                `json:"monthly_report"`
}

// DefaultSettings returns sensible defaults for a 1–3 chair clinic.
func DefaultSettings() Settings {
	day := DayHours{Open: "09:00", Close: "19:00"}
	return Settings{
		ReminderEveningHour:   20,
		TwoHourReminder:       true,
		QuietStartHour:        22,
		QuietEndHour:          8,
		Chairs:                1,
		SlotMinutes:           30,
		WorkingHours:          map[string]DayHours{"mon": day, "tue": day, "wed": day, "thu": day, "fri": day, "sat": {Open: "09:00", Close: "15:00"}},
		AutoBookFromSite:      false,
		AutoCompleteConfirmed: true,
		ReviewRequests:        true,
		WeeklyReport:          true,
		MonthlyReport:         true,
	}
}

// Normalize fills zero values with defaults and clamps ranges.
func (s *Settings) Normalize() {
	d := DefaultSettings()
	if s.ReminderEveningHour < 8 || s.ReminderEveningHour > 21 {
		s.ReminderEveningHour = d.ReminderEveningHour
	}
	if s.QuietStartHour <= 0 || s.QuietStartHour > 23 {
		s.QuietStartHour = d.QuietStartHour
	}
	if s.QuietEndHour < 0 || s.QuietEndHour > 12 {
		s.QuietEndHour = d.QuietEndHour
	}
	if s.Chairs <= 0 || s.Chairs > 50 {
		s.Chairs = d.Chairs
	}
	if s.SlotMinutes < 10 || s.SlotMinutes > 240 {
		s.SlotMinutes = d.SlotMinutes
	}
	if s.WorkingHours == nil {
		s.WorkingHours = d.WorkingHours
	}
}

// User is a dashboard user.
type User struct {
	ID           uuid.UUID  `json:"id"`
	ClinicID     *uuid.UUID `json:"clinic_id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Name         string     `json:"name"`
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`

	MustChangePassword bool       `json:"must_change_password"`
	TOTPEnabled        bool       `json:"totp_enabled"`
	TOTPSecretEnc      []byte     `json:"-"`
	TOTPLastStep       int64      `json:"-"`
	FailedLogins       int        `json:"-"`
	LockedUntil        *time.Time `json:"-"`
}

// Doctor works in a clinic.
type Doctor struct {
	ID        uuid.UUID `json:"id"`
	ClinicID  uuid.UUID `json:"clinic_id"`
	Name      string    `json:"name"`
	Specialty string    `json:"specialty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Service is a clinic's price-list item.
type Service struct {
	ID          uuid.UUID `json:"id"`
	ClinicID    uuid.UUID `json:"clinic_id"`
	Name        string    `json:"name"`
	DurationMin int       `json:"duration_min"`
	PriceKZT    int       `json:"price_kzt"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// Patient of a clinic. Phone is digits-only international format (77011234567).
type Patient struct {
	ID         uuid.UUID  `json:"id"`
	ClinicID   uuid.UUID  `json:"clinic_id"`
	Phone      string     `json:"phone"`
	Name       string     `json:"name"`
	Language   string     `json:"language"`
	ConsentAt  *time.Time `json:"consent_at"`
	OptedOutAt *time.Time `json:"opted_out_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// Appointment is a booked visit.
type Appointment struct {
	ID              uuid.UUID  `json:"id"`
	ClinicID        uuid.UUID  `json:"clinic_id"`
	PatientID       uuid.UUID  `json:"patient_id"`
	DoctorID        *uuid.UUID `json:"doctor_id"`
	ServiceID       *uuid.UUID `json:"service_id"`
	ServiceName     string     `json:"service_name"`
	StartsAt        time.Time  `json:"starts_at"`
	DurationMin     int        `json:"duration_min"`
	Status          string     `json:"status"`
	Source          string     `json:"source"`
	LeadID          *uuid.UUID `json:"lead_id"`
	Notes           string     `json:"notes"`
	ConfirmedAt     *time.Time `json:"confirmed_at"`
	StatusChangedAt time.Time  `json:"status_changed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Joined, read-only.
	PatientName  string `json:"patient_name,omitempty"`
	PatientPhone string `json:"patient_phone,omitempty"`
	DoctorName   string `json:"doctor_name,omitempty"`
}

// EndsAt returns the appointment end time.
func (a *Appointment) EndsAt() time.Time {
	return a.StartsAt.Add(time.Duration(a.DurationMin) * time.Minute)
}

// Lead is an incoming booking request (website, WhatsApp, 2GIS...).
type Lead struct {
	ID            uuid.UUID         `json:"id"`
	ClinicID      uuid.UUID         `json:"clinic_id"`
	Source        string            `json:"source"`
	Name          string            `json:"name"`
	Phone         string            `json:"phone"`
	Service       string            `json:"service"`
	PreferredAt   *time.Time        `json:"preferred_at"`
	Comment       string            `json:"comment"`
	Status        string            `json:"status"`
	PageURL       string            `json:"page_url"`
	UTM           map[string]string `json:"utm"`
	AppointmentID *uuid.UUID        `json:"appointment_id"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// Message is a logged WhatsApp message (inbound or outbound).
type Message struct {
	ID                uuid.UUID  `json:"id"`
	ClinicID          uuid.UUID  `json:"clinic_id"`
	PatientID         *uuid.UUID `json:"patient_id"`
	AppointmentID     *uuid.UUID `json:"appointment_id"`
	Direction         string     `json:"direction"`
	Phone             string     `json:"phone"`
	Kind              string     `json:"kind"`
	Body              string     `json:"body"`
	Status            string     `json:"status"`
	ProviderMessageID *string    `json:"provider_message_id"`
	Attempts          int        `json:"attempts"`
	LastError         string     `json:"last_error"`
	CreatedAt         time.Time  `json:"created_at"`
	SentAt            *time.Time `json:"sent_at"`
}

// Review is a post-visit rating.
type Review struct {
	ID            uuid.UUID  `json:"id"`
	ClinicID      uuid.UUID  `json:"clinic_id"`
	AppointmentID *uuid.UUID `json:"appointment_id"`
	PatientID     *uuid.UUID `json:"patient_id"`
	Rating        int        `json:"rating"`
	Comment       string     `json:"comment"`
	RoutedTo      string     `json:"routed_to"`
	CreatedAt     time.Time  `json:"created_at"`
}

// SiteKey authorises a clinic website to push bookings.
type SiteKey struct {
	ID             uuid.UUID  `json:"id"`
	ClinicID       uuid.UUID  `json:"clinic_id"`
	Name           string     `json:"name"`
	KeyPrefix      string     `json:"key_prefix"`
	AllowedOrigins []string   `json:"allowed_origins"`
	LastUsedAt     *time.Time `json:"last_used_at"`
	RevokedAt      *time.Time `json:"revoked_at"`
	CreatedAt      time.Time  `json:"created_at"`
}
