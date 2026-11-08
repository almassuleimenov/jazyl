// Shapes returned by the Jazyl API (backend/internal/domain, store/analytics.go).

export type Role = "superadmin" | "owner" | "admin";
export type Plan = "no_show" | "full_chair" | "growth";
export type ClinicStatus = "pilot" | "active" | "paused" | "churned";
export type ApptStatus = "scheduled" | "confirmed" | "reschedule_requested" | "cancelled" | "completed" | "no_show";
export type LeadStatus = "new" | "contacted" | "booked" | "rejected";

export type User = {
  id: string;
  clinic_id: string | null;
  email: string;
  name: string;
  role: Role;
  is_active: boolean;
  last_login_at: string | null;
  created_at: string;
  must_change_password: boolean;
  totp_enabled: boolean;
};

export type DayHours = { open: string; close: string };

export type ClinicSettings = {
  reminder_evening_hour: number;
  two_hour_reminder: boolean;
  quiet_start_hour: number;
  quiet_end_hour: number;
  chairs: number;
  slot_minutes: number;
  working_hours: Record<string, DayHours>;
  auto_book_from_site: boolean;
  auto_complete_confirmed: boolean;
  review_requests: boolean;
  weekly_report: boolean;
  monthly_report: boolean;
};

export type Clinic = {
  id: string;
  slug: string;
  name: string;
  city: string;
  address: string;
  timezone: string;
  language: string;
  plan: Plan;
  status: ClinicStatus;
  owner_phone: string;
  admin_phones: string[];
  gis_review_url: string;
  avg_check_kzt: number;
  settings: ClinicSettings;
  wa_provider: "openwa" | "mock";
  wa_session_id: string | null;
  wa_session_status: string;
  created_at: string;
  updated_at: string;
};

export type Me = {
  user: User;
  clinic?: Clinic;
  session: { mfa: boolean; mfa_setup_required: boolean; must_change_password: boolean };
};

export type Appointment = {
  id: string;
  clinic_id: string;
  patient_id: string;
  doctor_id: string | null;
  service_id: string | null;
  service_name: string;
  starts_at: string;
  duration_min: number;
  status: ApptStatus;
  source: string;
  notes: string;
  confirmed_at: string | null;
  status_changed_at: string;
  patient_name?: string;
  patient_phone?: string;
  doctor_name?: string;
};

export type Lead = {
  id: string;
  source: string;
  name: string;
  phone: string;
  service: string;
  preferred_at: string | null;
  comment: string;
  status: LeadStatus;
  page_url: string;
  appointment_id: string | null;
  created_at: string;
};

export type Message = {
  id: string;
  direction: "in" | "out";
  phone: string;
  kind: string;
  body: string;
  status: string;
  last_error: string;
  created_at: string;
  sent_at: string | null;
};

export type ChatSummary = {
  phone: string;
  patient_id: string | null;
  patient_name: string;
  in_count: number;
  out_count: number;
  last_message_at: string;
  last_body: string;
  last_direction: "in" | "out";
};

export type Metrics = {
  from: string;
  to: string;
  appointments: number;
  reminders_sent: number;
  confirmed: number;
  reschedule_requested: number;
  cancelled_by_patient: number;
  completed: number;
  no_shows: number;
  prevented_no_shows: number;
  saved_revenue_kzt: number;
  confirmation_rate: number;
  no_show_rate: number;
  leads: number;
  leads_booked: number;
  reviews: number;
  avg_rating: number;
  complaints: number;
  bot_chats: number;
};

export type PlatformRow = {
  clinic_id: string;
  name: string;
  city: string;
  plan: Plan;
  status: ClinicStatus;
  wa_session_status: string;
  appointments_30d: number;
  reminders_30d: number;
  failed_30d: number;
  last_inbound_at: string | null;
};

export type PlatformOverview = {
  clinics: PlatformRow[];
  total: number;
  paying: number;
  mrr_kzt: number;
  by_plan: Record<string, number>;
};

export type SiteKey = {
  id: string;
  name: string;
  key_prefix: string;
  allowed_origins: string[];
  last_used_at: string | null;
  revoked_at: string | null;
  created_at: string;
};

export type Doctor = { id: string; name: string; specialty: string; is_active: boolean };
export type Service = { id: string; name: string; duration_min: number; price_kzt: number; is_active: boolean };

export type Page<T> = { items: T[]; next_cursor?: string };
