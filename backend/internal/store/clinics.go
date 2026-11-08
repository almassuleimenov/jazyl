package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/almasik/jazyl/internal/domain"
)

const clinicCols = `id, slug, name, city, address, timezone, language, plan, status, owner_phone, admin_phones,
	gis_review_url, avg_check_kzt, settings, wa_provider, wa_session_id, wa_session_status, created_at, updated_at`

func scanClinic(row pgx.Row) (*domain.Clinic, error) {
	var c domain.Clinic
	var settings []byte
	err := row.Scan(&c.ID, &c.Slug, &c.Name, &c.City, &c.Address, &c.Timezone, &c.Language, &c.Plan, &c.Status,
		&c.OwnerPhone, &c.AdminPhones, &c.GisReviewURL, &c.AvgCheckKZT, &settings, &c.WAProvider,
		&c.WASessionID, &c.WASessionStatus, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	c.Settings = domain.DefaultSettings()
	if len(settings) > 0 {
		if err := json.Unmarshal(settings, &c.Settings); err != nil {
			return nil, fmt.Errorf("decode clinic settings: %w", err)
		}
	}
	c.Settings.Normalize()
	if c.AdminPhones == nil {
		c.AdminPhones = []string{}
	}
	return &c, nil
}

// CreateClinic inserts a clinic.
func (s *Store) CreateClinic(ctx context.Context, c *domain.Clinic) (*domain.Clinic, error) {
	settings, err := json.Marshal(c.Settings)
	if err != nil {
		return nil, fmt.Errorf("encode settings: %w", err)
	}
	return scanClinic(s.db.QueryRow(ctx, `
		INSERT INTO clinics (slug, name, city, address, timezone, language, plan, status, owner_phone, admin_phones,
			gis_review_url, avg_check_kzt, settings, wa_provider)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING `+clinicCols,
		c.Slug, c.Name, c.City, c.Address, c.Timezone, c.Language, c.Plan, c.Status, c.OwnerPhone, c.AdminPhones,
		c.GisReviewURL, c.AvgCheckKZT, settings, c.WAProvider))
}

// UpdateClinic overwrites editable clinic fields.
func (s *Store) UpdateClinic(ctx context.Context, c *domain.Clinic) (*domain.Clinic, error) {
	settings, err := json.Marshal(c.Settings)
	if err != nil {
		return nil, fmt.Errorf("encode settings: %w", err)
	}
	return scanClinic(s.db.QueryRow(ctx, `
		UPDATE clinics SET name=$2, city=$3, address=$4, timezone=$5, language=$6, plan=$7, status=$8,
			owner_phone=$9, admin_phones=$10, gis_review_url=$11, avg_check_kzt=$12, settings=$13, wa_provider=$14,
			updated_at=now()
		WHERE id=$1
		RETURNING `+clinicCols,
		c.ID, c.Name, c.City, c.Address, c.Timezone, c.Language, c.Plan, c.Status, c.OwnerPhone, c.AdminPhones,
		c.GisReviewURL, c.AvgCheckKZT, settings, c.WAProvider))
}

// GetClinic returns a clinic by id.
func (s *Store) GetClinic(ctx context.Context, id uuid.UUID) (*domain.Clinic, error) {
	return scanClinic(s.db.QueryRow(ctx, `SELECT `+clinicCols+` FROM clinics WHERE id=$1`, id))
}

// GetClinicBySession returns the clinic linked to an OpenWA session.
func (s *Store) GetClinicBySession(ctx context.Context, sessionID string) (*domain.Clinic, error) {
	return scanClinic(s.db.QueryRow(ctx, `SELECT `+clinicCols+` FROM clinics WHERE wa_session_id=$1`, sessionID))
}

// ListClinics returns all clinics (superadmin view), newest first.
func (s *Store) ListClinics(ctx context.Context) ([]*domain.Clinic, error) {
	return s.queryClinics(ctx, `SELECT `+clinicCols+` FROM clinics ORDER BY created_at DESC`)
}

// ListOperatingClinics returns clinics whose automation should run.
func (s *Store) ListOperatingClinics(ctx context.Context) ([]*domain.Clinic, error) {
	return s.queryClinics(ctx, `SELECT `+clinicCols+` FROM clinics WHERE status IN ('pilot','active') ORDER BY created_at`)
}

func (s *Store) queryClinics(ctx context.Context, q string, args ...any) ([]*domain.Clinic, error) {
	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	var out []*domain.Clinic
	for rows.Next() {
		c, err := scanClinic(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, mapErr(rows.Err())
}

// SetClinicSession links an OpenWA session to a clinic.
func (s *Store) SetClinicSession(ctx context.Context, clinicID uuid.UUID, sessionID *string, status string) error {
	tag, err := s.db.Exec(ctx, `UPDATE clinics SET wa_session_id=$2, wa_session_status=$3, updated_at=now() WHERE id=$1`,
		clinicID, sessionID, status)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// SetSessionStatus updates the cached WhatsApp session status by session id.
func (s *Store) SetSessionStatus(ctx context.Context, sessionID, status string) error {
	_, err := s.db.Exec(ctx, `UPDATE clinics SET wa_session_status=$2, updated_at=now() WHERE wa_session_id=$1`, sessionID, status)
	return mapErr(err)
}

// ---------- users ----------

const userCols = `id, clinic_id, email, password_hash, name, role, is_active, last_login_at, created_at,
	must_change_password, totp_enabled, totp_secret_enc, totp_last_step, failed_logins, locked_until`

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	if err := row.Scan(&u.ID, &u.ClinicID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt,
		&u.MustChangePassword, &u.TOTPEnabled, &u.TOTPSecretEnc, &u.TOTPLastStep, &u.FailedLogins, &u.LockedUntil); err != nil {
		return nil, mapErr(err)
	}
	return &u, nil
}

// CreateUser inserts a user.
func (s *Store) CreateUser(ctx context.Context, u *domain.User) (*domain.User, error) {
	return scanUser(s.db.QueryRow(ctx, `
		INSERT INTO users (clinic_id, email, password_hash, name, role, must_change_password)
		VALUES ($1, lower($2), $3, $4, $5, $6)
		RETURNING `+userCols, u.ClinicID, u.Email, u.PasswordHash, u.Name, u.Role, u.MustChangePassword))
}

// GetUserByEmail looks up a user case-insensitively.
func (s *Store) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return scanUser(s.db.QueryRow(ctx, `SELECT `+userCols+` FROM users WHERE lower(email)=lower($1)`, email))
}

// GetUser returns a user by id.
func (s *Store) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return scanUser(s.db.QueryRow(ctx, `SELECT `+userCols+` FROM users WHERE id=$1`, id))
}

// ListClinicUsers lists users of a clinic.
func (s *Store) ListClinicUsers(ctx context.Context, clinicID uuid.UUID) ([]*domain.User, error) {
	rows, err := s.db.Query(ctx, `SELECT `+userCols+` FROM users WHERE clinic_id=$1 ORDER BY created_at`, clinicID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	var out []*domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, mapErr(rows.Err())
}

// SetUserActive enables/disables a clinic user.
func (s *Store) SetUserActive(ctx context.Context, clinicID, userID uuid.UUID, active bool) error {
	tag, err := s.db.Exec(ctx, `UPDATE users SET is_active=$3 WHERE id=$2 AND clinic_id=$1`, clinicID, userID, active)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// SetUserPassword replaces a user's password hash. mustChange forces a change at next login
// (set when someone else chose the password). It also clears any lockout.
func (s *Store) SetUserPassword(ctx context.Context, userID uuid.UUID, hash string, mustChange bool) error {
	tag, err := s.db.Exec(ctx, `UPDATE users SET password_hash=$2, must_change_password=$3, password_changed_at=now(),
		failed_logins=0, locked_until=NULL WHERE id=$1`, userID, hash, mustChange)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// TouchLogin records a successful login and clears the failure counter.
func (s *Store) TouchLogin(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `UPDATE users SET last_login_at=now(), failed_logins=0, locked_until=NULL WHERE id=$1`, userID)
	return mapErr(err)
}

// RecordLoginFailure counts a failed password or code and locks the account for lockFor
// once maxFailures is reached. It returns the lock deadline when this failure locked the account.
func (s *Store) RecordLoginFailure(ctx context.Context, userID uuid.UUID, maxFailures int, lockFor time.Duration) (*time.Time, error) {
	var locked *time.Time
	var justLocked bool
	err := s.db.QueryRow(ctx, `
		UPDATE users SET
			failed_logins = CASE WHEN failed_logins + 1 >= $2 THEN 0 ELSE failed_logins + 1 END,
			locked_until  = CASE WHEN failed_logins + 1 >= $2 THEN now() + make_interval(secs => $3) ELSE locked_until END
		WHERE id=$1
		RETURNING locked_until, COALESCE(locked_until > now(), false)`, userID, maxFailures, lockFor.Seconds()).Scan(&locked, &justLocked)
	if err != nil {
		return nil, mapErr(err)
	}
	if !justLocked {
		return nil, nil
	}
	return locked, nil
}

// SetPendingTOTP stores a new, not yet enabled, encrypted TOTP secret.
func (s *Store) SetPendingTOTP(ctx context.Context, userID uuid.UUID, sealed []byte) error {
	tag, err := s.db.Exec(ctx, `UPDATE users SET totp_secret_enc=$2 WHERE id=$1 AND NOT totp_enabled`, userID, sealed)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConflict
	}
	return nil
}

// EnableTOTP turns on 2FA after the first code was verified at step.
func (s *Store) EnableTOTP(ctx context.Context, userID uuid.UUID, step int64) error {
	tag, err := s.db.Exec(ctx, `UPDATE users SET totp_enabled=true, totp_last_step=$2
		WHERE id=$1 AND NOT totp_enabled AND totp_secret_enc IS NOT NULL`, userID, step)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConflict
	}
	return nil
}

// DisableTOTP removes the second factor (the user's own request or a reset by Jazyl).
func (s *Store) DisableTOTP(ctx context.Context, userID uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `UPDATE users SET totp_enabled=false, totp_secret_enc=NULL, totp_last_step=0 WHERE id=$1`, userID)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ConsumeTOTPStep records step as used. It fails with ErrConflict if this step, or a later one,
// was already accepted, so each code works exactly once.
func (s *Store) ConsumeTOTPStep(ctx context.Context, userID uuid.UUID, step int64) error {
	tag, err := s.db.Exec(ctx, `UPDATE users SET totp_last_step=$2 WHERE id=$1 AND totp_last_step < $2`, userID, step)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConflict
	}
	return nil
}

// GetClinicUser returns a user only if it belongs to clinicID.
func (s *Store) GetClinicUser(ctx context.Context, clinicID, userID uuid.UUID) (*domain.User, error) {
	return scanUser(s.db.QueryRow(ctx, `SELECT `+userCols+` FROM users WHERE id=$1 AND clinic_id=$2`, userID, clinicID))
}

// CountSuperadmins counts platform admins (bootstrap check).
func (s *Store) CountSuperadmins(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRow(ctx, `SELECT count(*) FROM users WHERE role='superadmin'`).Scan(&n)
	return n, mapErr(err)
}

// ---------- refresh tokens ----------

// InsertRefreshToken stores a hashed refresh token. mfa records that the session passed 2FA.
func (s *Store) InsertRefreshToken(ctx context.Context, userID uuid.UUID, hash []byte, expires time.Time, ua, ip string, mfa bool) error {
	_, err := s.db.Exec(ctx, `INSERT INTO refresh_tokens (user_id, token_hash, expires_at, user_agent, ip, mfa) VALUES ($1,$2,$3,$4,$5,$6)`,
		userID, hash, expires, truncate(ua, 300), ip, mfa)
	return mapErr(err)
}

// RefreshReuseGrace is how long a just-rotated refresh token may be presented again without
// being treated as stolen. Parallel requests from one browser race on the same token; the
// loser gets ErrRefreshRace (a plain 401) instead of revoking every session of the user.
const RefreshReuseGrace = 30 * time.Second

// ErrRefreshRace marks a benign reuse of a token rotated moments ago.
var ErrRefreshRace = fmt.Errorf("%w: refresh token was just rotated", domain.ErrUnauthorized)

// ConsumeRefreshToken atomically revokes a valid token and returns its owner.
// A token that was already revoked signals reuse: all of the user's tokens are revoked,
// except within RefreshReuseGrace of a rotation (see ErrRefreshRace).
// The returned mfa flag tells whether that session had passed the second factor.
func (s *Store) ConsumeRefreshToken(ctx context.Context, hash []byte) (userID uuid.UUID, mfa bool, err error) {
	var revoked, rotated *time.Time
	var expires time.Time
	err = s.db.QueryRow(ctx, `SELECT user_id, revoked_at, rotated_at, expires_at, mfa FROM refresh_tokens WHERE token_hash=$1 FOR UPDATE`, hash).
		Scan(&userID, &revoked, &rotated, &expires, &mfa)
	if err != nil {
		return uuid.Nil, false, mapErr(err)
	}
	if revoked != nil && rotated != nil && time.Since(*rotated) < RefreshReuseGrace {
		return uuid.Nil, false, ErrRefreshRace
	}
	if revoked != nil {
		if _, err := s.db.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=now() WHERE user_id=$1 AND revoked_at IS NULL`, userID); err != nil {
			return uuid.Nil, false, mapErr(err)
		}
		return uuid.Nil, false, domain.ErrUnauthorized
	}
	if time.Now().After(expires) {
		return uuid.Nil, false, domain.ErrUnauthorized
	}
	if _, err := s.db.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=now(), rotated_at=now() WHERE token_hash=$1`, hash); err != nil {
		return uuid.Nil, false, mapErr(err)
	}
	return userID, mfa, nil
}

// RevokeRefreshToken revokes a single token (logout).
func (s *Store) RevokeRefreshToken(ctx context.Context, hash []byte) error {
	_, err := s.db.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=now() WHERE token_hash=$1 AND revoked_at IS NULL`, hash)
	return mapErr(err)
}

// RevokeUserTokens revokes every token of a user (password change, deactivation).
func (s *Store) RevokeUserTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=now() WHERE user_id=$1 AND revoked_at IS NULL`, userID)
	return mapErr(err)
}

// ---------- site keys ----------

const siteKeyCols = `id, clinic_id, name, key_prefix, allowed_origins, last_used_at, revoked_at, created_at`

func scanSiteKey(row pgx.Row) (*domain.SiteKey, error) {
	var k domain.SiteKey
	if err := row.Scan(&k.ID, &k.ClinicID, &k.Name, &k.KeyPrefix, &k.AllowedOrigins, &k.LastUsedAt, &k.RevokedAt, &k.CreatedAt); err != nil {
		return nil, mapErr(err)
	}
	if k.AllowedOrigins == nil {
		k.AllowedOrigins = []string{}
	}
	return &k, nil
}

// CreateSiteKey stores a hashed site key.
func (s *Store) CreateSiteKey(ctx context.Context, clinicID uuid.UUID, name, prefix string, hash []byte, origins []string) (*domain.SiteKey, error) {
	if origins == nil {
		origins = []string{}
	}
	return scanSiteKey(s.db.QueryRow(ctx, `
		INSERT INTO site_keys (clinic_id, name, key_prefix, key_hash, allowed_origins) VALUES ($1,$2,$3,$4,$5)
		RETURNING `+siteKeyCols, clinicID, name, prefix, hash, origins))
}

// ListSiteKeys lists a clinic's site keys.
func (s *Store) ListSiteKeys(ctx context.Context, clinicID uuid.UUID) ([]*domain.SiteKey, error) {
	rows, err := s.db.Query(ctx, `SELECT `+siteKeyCols+` FROM site_keys WHERE clinic_id=$1 ORDER BY created_at DESC`, clinicID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	var out []*domain.SiteKey
	for rows.Next() {
		k, err := scanSiteKey(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, mapErr(rows.Err())
}

// RevokeSiteKey revokes a site key.
func (s *Store) RevokeSiteKey(ctx context.Context, clinicID, keyID uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `UPDATE site_keys SET revoked_at=now() WHERE id=$2 AND clinic_id=$1 AND revoked_at IS NULL`, clinicID, keyID)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// GetActiveSiteKey resolves a key hash to an active key.
func (s *Store) GetActiveSiteKey(ctx context.Context, hash []byte) (*domain.SiteKey, error) {
	k, err := scanSiteKey(s.db.QueryRow(ctx, `SELECT `+siteKeyCols+` FROM site_keys WHERE key_hash=$1 AND revoked_at IS NULL`, hash))
	if err != nil {
		return nil, err
	}
	// Throttled last-used update; failures are irrelevant to the request.
	_, _ = s.db.Exec(ctx, `UPDATE site_keys SET last_used_at=now() WHERE id=$1 AND (last_used_at IS NULL OR last_used_at < now() - interval '5 minutes')`, k.ID)
	return k, nil
}

// ---------- audit ----------

// Audit writes an audit-log row. Errors are returned but callers usually only log them.
func (s *Store) Audit(ctx context.Context, clinicID, userID *uuid.UUID, action, entity, entityID, ip string, meta map[string]any) error {
	if meta == nil {
		meta = map[string]any{}
	}
	b, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("encode audit meta: %w", err)
	}
	_, err = s.db.Exec(ctx, `INSERT INTO audit_log (clinic_id, user_id, action, entity, entity_id, meta, ip) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		clinicID, userID, action, entity, entityID, b, ip)
	return mapErr(err)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
