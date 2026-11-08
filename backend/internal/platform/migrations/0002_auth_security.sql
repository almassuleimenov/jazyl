-- 0002: dashboard login security — TOTP 2FA, forced password change, account lockout.

ALTER TABLE users
    ADD COLUMN must_change_password boolean     NOT NULL DEFAULT false,
    ADD COLUMN totp_secret_enc      bytea,                 -- AES-256-GCM(nonce || ciphertext); NULL = no secret
    ADD COLUMN totp_enabled         boolean     NOT NULL DEFAULT false,
    ADD COLUMN totp_last_step       bigint      NOT NULL DEFAULT 0, -- replay protection: last accepted 30s step
    ADD COLUMN failed_logins        integer     NOT NULL DEFAULT 0,
    ADD COLUMN locked_until         timestamptz,
    ADD COLUMN password_changed_at  timestamptz;

-- Whether the session behind a refresh token passed the second factor; carried across rotations.
ALTER TABLE refresh_tokens
    ADD COLUMN mfa boolean NOT NULL DEFAULT false;
