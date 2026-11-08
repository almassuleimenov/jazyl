-- Jazyl core schema. Every tenant-owned row carries clinic_id; every query is scoped by it.

CREATE TABLE clinics (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug              text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$'),
    name              text NOT NULL,
    city              text NOT NULL DEFAULT '',
    address           text NOT NULL DEFAULT '',
    timezone          text NOT NULL DEFAULT 'Asia/Almaty',
    language          text NOT NULL DEFAULT 'ru' CHECK (language IN ('ru', 'kk')),
    plan              text NOT NULL DEFAULT 'no_show' CHECK (plan IN ('no_show', 'full_chair', 'growth')),
    status            text NOT NULL DEFAULT 'pilot' CHECK (status IN ('pilot', 'active', 'paused', 'churned')),
    owner_phone       text NOT NULL DEFAULT '',
    admin_phones      text[] NOT NULL DEFAULT '{}',
    gis_review_url    text NOT NULL DEFAULT '',
    avg_check_kzt     integer NOT NULL DEFAULT 15000 CHECK (avg_check_kzt >= 0),
    settings          jsonb NOT NULL DEFAULT '{}'::jsonb,
    wa_provider       text NOT NULL DEFAULT 'openwa' CHECK (wa_provider IN ('openwa', 'mock')),
    wa_session_id     text UNIQUE,
    wa_session_status text NOT NULL DEFAULT 'none',
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id     uuid REFERENCES clinics(id) ON DELETE CASCADE,
    email         text NOT NULL,
    password_hash text NOT NULL,
    name          text NOT NULL DEFAULT '',
    role          text NOT NULL CHECK (role IN ('superadmin', 'owner', 'admin')),
    is_active     boolean NOT NULL DEFAULT true,
    last_login_at timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_role_scope CHECK ((role = 'superadmin') = (clinic_id IS NULL))
);
CREATE UNIQUE INDEX users_email_uniq ON users (lower(email));
CREATE INDEX users_clinic_idx ON users (clinic_id);

CREATE TABLE refresh_tokens (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    user_agent text NOT NULL DEFAULT '',
    ip         text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX refresh_tokens_user_idx ON refresh_tokens (user_id);

-- Keys used by clinic websites (hosted separately, e.g. on Vercel) to push bookings.
CREATE TABLE site_keys (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id       uuid NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    name            text NOT NULL DEFAULT '',
    key_prefix      text NOT NULL,
    key_hash        bytea NOT NULL UNIQUE,
    allowed_origins text[] NOT NULL DEFAULT '{}',
    last_used_at    timestamptz,
    revoked_at      timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX site_keys_clinic_idx ON site_keys (clinic_id);

CREATE TABLE doctors (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id  uuid NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    name       text NOT NULL,
    specialty  text NOT NULL DEFAULT '',
    is_active  boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX doctors_clinic_idx ON doctors (clinic_id);

CREATE TABLE services (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id    uuid NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    name         text NOT NULL,
    duration_min integer NOT NULL DEFAULT 30 CHECK (duration_min BETWEEN 5 AND 480),
    price_kzt    integer NOT NULL DEFAULT 0 CHECK (price_kzt >= 0),
    is_active    boolean NOT NULL DEFAULT true,
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX services_clinic_idx ON services (clinic_id);

CREATE TABLE patients (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id    uuid NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    phone        text NOT NULL CHECK (phone ~ '^[0-9]{10,15}$'),
    name         text NOT NULL DEFAULT '',
    language     text NOT NULL DEFAULT 'ru' CHECK (language IN ('ru', 'kk')),
    consent_at   timestamptz,
    opted_out_at timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (clinic_id, phone)
);

CREATE TABLE leads (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id      uuid NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    source         text NOT NULL CHECK (source IN ('site', 'whatsapp', '2gis', 'instagram', 'other')),
    name           text NOT NULL DEFAULT '',
    phone          text NOT NULL,
    service        text NOT NULL DEFAULT '',
    preferred_at   timestamptz,
    comment        text NOT NULL DEFAULT '',
    status         text NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'contacted', 'booked', 'rejected')),
    page_url       text NOT NULL DEFAULT '',
    utm            jsonb NOT NULL DEFAULT '{}'::jsonb,
    appointment_id uuid,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX leads_clinic_created_idx ON leads (clinic_id, created_at DESC);

CREATE TABLE appointments (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id         uuid NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    patient_id        uuid NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    doctor_id         uuid REFERENCES doctors(id) ON DELETE SET NULL,
    service_id        uuid REFERENCES services(id) ON DELETE SET NULL,
    service_name      text NOT NULL DEFAULT '',
    starts_at         timestamptz NOT NULL,
    duration_min      integer NOT NULL DEFAULT 30 CHECK (duration_min BETWEEN 5 AND 480),
    status            text NOT NULL DEFAULT 'scheduled'
                      CHECK (status IN ('scheduled', 'confirmed', 'reschedule_requested', 'cancelled', 'completed', 'no_show')),
    source            text NOT NULL DEFAULT 'manual' CHECK (source IN ('manual', 'notebook', 'site', 'crm', 'whatsapp')),
    lead_id           uuid REFERENCES leads(id) ON DELETE SET NULL,
    notes             text NOT NULL DEFAULT '',
    confirmed_at      timestamptz,
    status_changed_at timestamptz NOT NULL DEFAULT now(),
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX appointments_clinic_starts_idx ON appointments (clinic_id, starts_at);
CREATE INDEX appointments_patient_idx ON appointments (patient_id, starts_at DESC);
-- Used by the auto-complete sweep.
CREATE INDEX appointments_open_idx ON appointments (starts_at) WHERE status IN ('scheduled', 'confirmed');
-- Same patient cannot be booked twice for the same minute in a clinic (idempotent notebook imports).
CREATE UNIQUE INDEX appointments_dedupe_uniq ON appointments (clinic_id, patient_id, starts_at)
    WHERE status NOT IN ('cancelled');

ALTER TABLE leads ADD CONSTRAINT leads_appointment_fk
    FOREIGN KEY (appointment_id) REFERENCES appointments(id) ON DELETE SET NULL;

-- Outbound + inbound WhatsApp message log. Outbound rows double as a transactional outbox.
CREATE TABLE messages (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id           uuid NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    patient_id          uuid REFERENCES patients(id) ON DELETE SET NULL,
    appointment_id      uuid REFERENCES appointments(id) ON DELETE SET NULL,
    direction           text NOT NULL CHECK (direction IN ('in', 'out')),
    phone               text NOT NULL,
    kind                text NOT NULL,
    body                text NOT NULL,
    status              text NOT NULL CHECK (status IN ('queued', 'sending', 'sent', 'failed', 'received', 'cancelled')),
    provider_message_id text,
    attempts            integer NOT NULL DEFAULT 0,
    next_attempt_at     timestamptz NOT NULL DEFAULT now(),
    locked_until        timestamptz,
    last_error          text NOT NULL DEFAULT '',
    created_at          timestamptz NOT NULL DEFAULT now(),
    sent_at             timestamptz
);
CREATE INDEX messages_outbox_idx ON messages (next_attempt_at) WHERE status = 'queued';
CREATE INDEX messages_stuck_idx ON messages (locked_until) WHERE status = 'sending';
CREATE INDEX messages_clinic_created_idx ON messages (clinic_id, created_at DESC);
CREATE INDEX messages_clinic_phone_idx ON messages (clinic_id, phone, created_at DESC);

CREATE TABLE reminders (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id      uuid NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    appointment_id uuid NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    kind           text NOT NULL CHECK (kind IN ('day_before', 'two_hours', 'review_request')),
    send_at        timestamptz NOT NULL,
    status         text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'skipped')),
    skip_reason    text NOT NULL DEFAULT '',
    message_id     uuid REFERENCES messages(id) ON DELETE SET NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    processed_at   timestamptz,
    UNIQUE (appointment_id, kind)
);
CREATE INDEX reminders_due_idx ON reminders (send_at) WHERE status = 'pending';
CREATE INDEX reminders_clinic_idx ON reminders (clinic_id, send_at);

-- What answer the bot is waiting for from a given phone (button-like "1 / 2" flows, no free chat).
CREATE TABLE conversations (
    clinic_id      uuid NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    phone          text NOT NULL,
    awaiting       text NOT NULL CHECK (awaiting IN ('reminder_reply', 'review_rating', 'review_comment')),
    appointment_id uuid REFERENCES appointments(id) ON DELETE CASCADE,
    expires_at     timestamptz NOT NULL,
    updated_at     timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (clinic_id, phone)
);

CREATE TABLE reviews (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id      uuid NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    appointment_id uuid UNIQUE REFERENCES appointments(id) ON DELETE SET NULL,
    patient_id     uuid REFERENCES patients(id) ON DELETE SET NULL,
    rating         smallint NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment        text NOT NULL DEFAULT '',
    routed_to      text NOT NULL CHECK (routed_to IN ('2gis', 'owner')),
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX reviews_clinic_created_idx ON reviews (clinic_id, created_at DESC);

-- "Notebook mode": the administrator sends tomorrow's list to the bot as plain text.
CREATE TABLE schedule_imports (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id    uuid NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    from_phone   text NOT NULL DEFAULT '',
    raw_text     text NOT NULL,
    target_date  date NOT NULL,
    parsed_count integer NOT NULL DEFAULT 0,
    failed_lines jsonb NOT NULL DEFAULT '[]'::jsonb,
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX schedule_imports_clinic_idx ON schedule_imports (clinic_id, created_at DESC);

CREATE TABLE reports (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id    uuid NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    period       text NOT NULL CHECK (period IN ('week', 'month')),
    period_start date NOT NULL,
    period_end   date NOT NULL,
    metrics      jsonb NOT NULL,
    message_id   uuid REFERENCES messages(id) ON DELETE SET NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (clinic_id, period, period_start)
);

-- Webhook de-duplication (OpenWA delivers at-least-once).
CREATE TABLE webhook_events (
    idempotency_key text PRIMARY KEY,
    received_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX webhook_events_received_idx ON webhook_events (received_at);

CREATE TABLE audit_log (
    id         bigserial PRIMARY KEY,
    clinic_id  uuid REFERENCES clinics(id) ON DELETE CASCADE,
    user_id    uuid REFERENCES users(id) ON DELETE SET NULL,
    action     text NOT NULL,
    entity     text NOT NULL DEFAULT '',
    entity_id  text NOT NULL DEFAULT '',
    meta       jsonb NOT NULL DEFAULT '{}'::jsonb,
    ip         text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX audit_log_clinic_idx ON audit_log (clinic_id, created_at DESC);

-- Status history: source of truth for analytics ("confirmed by patient", "reschedule instead of no-show").
CREATE TABLE appointment_events (
    id             bigserial PRIMARY KEY,
    clinic_id      uuid NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    appointment_id uuid NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    from_status    text NOT NULL DEFAULT '',
    to_status      text NOT NULL,
    actor          text NOT NULL CHECK (actor IN ('patient', 'staff', 'system')),
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX appointment_events_clinic_idx ON appointment_events (clinic_id, created_at);
CREATE INDEX appointment_events_appt_idx ON appointment_events (appointment_id);
