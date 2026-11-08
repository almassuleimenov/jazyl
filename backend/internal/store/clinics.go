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

const userCols = `id, clinic_id, email, password_hash, name, role, is_active, last_login_at, created_at`

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	if err := row.Scan(&u.ID, &u.ClinicID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt); err != nil {
		return nil, mapErr(err)
	}
	return &u, nil
}

// CreateUser inserts a user.
func (s *Store) CreateUser(ctx context.Context, u *domain.User) (*domain.User, error) {
	return scanUser(s.db.QueryRow(ctx, `
		INSERT INTO users (clinic_id, email, password_hash, name, role) VALUES ($1, lower($2), $3, $4, $5)
		RETURNING `+userCols, u.ClinicID, u.Email, u.PasswordHash, u.Name, u.Role))
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

// SetUserPassword replaces a user's password hash.
func (s *Store) SetUserPassword(ctx context.Context, userID uuid.UUID, hash string) error {
	_, err := s.db.Exec(ctx, `UPDATE users SET password_hash=$2 WHERE id=$1`, userID, hash)
	return mapErr(err)
}

// TouchLogin records a successful login.
func (s *Store) TouchLogin(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `UPDATE users SET last_login_at=now() WHERE id=$1`, userID)
	return mapErr(err)
}

// CountSuperadmins counts platform admins (bootstrap check).
func (s *Store) CountSuperadmins(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRow(ctx, `SELECT count(*) FROM users WHERE role='superadmin'`).Scan(&n)
	return n, mapErr(err)
}

// ---------- refresh tokens ----------

// InsertRefreshToken stores a hashed refresh token.
func (s *Store) InsertRefreshToken(ctx context.Context, userID uuid.UUID, hash []byte, expires time.Time, ua, ip string) error {
	_, err := s.db.Exec(ctx, `INSERT INTO refresh_tokens (user_id, token_hash, expires_at, user_agent, ip) VALUES ($1,$2,$3,$4,$5)`,
		userID, hash, expires, truncate(ua, 300), ip)
	return mapErr(err)
}

// ConsumeRefreshToken atomically revokes a valid token and returns its owner.
// A token that was already revoked signals reuse: all of the user's tokens are revoked.
func (s *Store) ConsumeRefreshToken(ctx context.Context, hash []byte) (uuid.UUID, error) {
	var userID uuid.UUID
	var revoked *time.Time
	var expires time.Time
	err := s.db.QueryRow(ctx, `SELECT user_id, revoked_at, expires_at FROM refresh_tokens WHERE token_hash=$1 FOR UPDATE`, hash).
		Scan(&userID, &revoked, &expires)
	if err != nil {
		return uuid.Nil, mapErr(err)
	}
	if revoked != nil {
		if _, err := s.db.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=now() WHERE user_id=$1 AND revoked_at IS NULL`, userID); err != nil {
			return uuid.Nil, mapErr(err)
		}
		return uuid.Nil, domain.ErrUnauthorized
	}
	if time.Now().After(expires) {
		return uuid.Nil, domain.ErrUnauthorized
	}
	if _, err := s.db.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=now() WHERE token_hash=$1`, hash); err != nil {
		return uuid.Nil, mapErr(err)
	}
	return userID, nil
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
