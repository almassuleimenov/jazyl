package api

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/almasik/jazyl/internal/auth"
	"github.com/almasik/jazyl/internal/domain"
	"github.com/almasik/jazyl/internal/realtime"
	"github.com/almasik/jazyl/internal/store"
)

// Login hardening. The per-IP rate limiter slows a single source; the per-account lockout
// stops distributed guessing against one account (password or TOTP code).
const (
	maxLoginFailures = 5
	lockoutDuration  = 15 * time.Minute
	mfaTokenTTL      = 5 * time.Minute
	totpIssuer       = "Jazyl"
)

type tokenResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	User         *domain.User `json:"user"`
	// Session flags the dashboard uses to route the user (password change, 2FA setup).
	MFA                bool `json:"mfa"`
	MFARequired        bool `json:"mfa_setup_required"`
	MustChangePassword bool `json:"must_change_password"`
}

func (a *API) issueTokens(r *http.Request, tx *store.Store, u *domain.User, mfa bool) (*tokenResponse, error) {
	p := auth.Principal{UserID: u.ID, Role: u.Role, ClinicID: u.ClinicID, MFA: mfa, MustChangePassword: u.MustChangePassword}
	access, err := a.tokens.Issue(p, "", a.cfg.AccessTokenTTL)
	if err != nil {
		return nil, err
	}
	refresh, hash, err := auth.RandomToken("jzr_", 32)
	if err != nil {
		return nil, err
	}
	if err := tx.InsertRefreshToken(r.Context(), u.ID, hash, time.Now().Add(a.cfg.RefreshTokenTTL), r.UserAgent(), a.clientIP(r), mfa); err != nil {
		return nil, err
	}
	return &tokenResponse{
		AccessToken: access, RefreshToken: refresh, TokenType: "Bearer",
		ExpiresIn: int(a.cfg.AccessTokenTTL.Seconds()), User: u,
		MFA: mfa, MFARequired: a.mfaGateClosed(p), MustChangePassword: u.MustChangePassword,
	}, nil
}

// mfaGateClosed reports whether p is a superadmin session that still has to pass 2FA.
func (a *API) mfaGateClosed(p auth.Principal) bool {
	return a.cfg.RequireSuperadminMFA && p.Role == domain.RoleSuperadmin && !p.MFA
}

func locked(u *domain.User) bool {
	return u != nil && u.LockedUntil != nil && u.LockedUntil.After(time.Now())
}

func writeLocked(w http.ResponseWriter, r *http.Request, until time.Time) {
	secs := int(math.Ceil(time.Until(until).Seconds()))
	if secs < 1 {
		secs = 1
	}
	w.Header().Set("Retry-After", fmt.Sprint(secs))
	writeErr(w, r, http.StatusTooManyRequests, "account_locked",
		fmt.Sprintf("too many failed attempts; try again in %d min", (secs+59)/60))
}

// recordFailure counts a failed password/code. It returns true (and writes the response)
// when this failure locked the account.
func (a *API) recordFailure(w http.ResponseWriter, r *http.Request, u *domain.User, reason string) bool {
	until, err := a.store.RecordLoginFailure(r.Context(), u.ID, maxLoginFailures, lockoutDuration)
	if err != nil {
		logger(r.Context()).Warn("record login failure", "err", err)
		return false
	}
	logger(r.Context()).Warn("auth failure", "reason", reason, "user_id", u.ID, "ip", a.clientIP(r))
	if until != nil {
		a.audit(r, u.ClinicID, "auth.locked", "user", u.ID.String(), map[string]any{"reason": reason})
		writeLocked(w, r, *until)
		return true
	}
	return false
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decode(w, r, &in) {
		return
	}
	u, err := a.store.GetUserByEmail(r.Context(), strings.TrimSpace(in.Email))
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		fail(w, r, err)
		return
	}
	if locked(u) {
		auth.CheckPassword("", in.Password) // keep timing uniform; do not test the real hash while locked
		writeLocked(w, r, *u.LockedUntil)
		return
	}
	hash := ""
	if u != nil {
		hash = u.PasswordHash
	}
	if !auth.CheckPassword(hash, in.Password) || u == nil || !u.IsActive {
		if u != nil && u.IsActive && a.recordFailure(w, r, u, "password") {
			return
		}
		logger(r.Context()).Warn("login failed", "email_domain", emailDomain(in.Email), "ip", a.clientIP(r))
		writeErr(w, r, http.StatusUnauthorized, "invalid_credentials", "wrong email or password")
		return
	}
	if u.TOTPEnabled {
		// Password is right; the session starts only after the TOTP code (POST /auth/mfa).
		t, err := a.tokens.Issue(auth.Principal{UserID: u.ID, Role: u.Role, ClinicID: u.ClinicID}, "mfa", mfaTokenTTL)
		if err != nil {
			fail(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"mfa_required": true, "mfa_token": t, "expires_in": int(mfaTokenTTL.Seconds())})
		return
	}
	a.completeLogin(w, r, u, false)
}

func (a *API) completeLogin(w http.ResponseWriter, r *http.Request, u *domain.User, mfa bool) {
	resp, err := a.issueTokens(r, a.store, u, mfa)
	if err != nil {
		fail(w, r, err)
		return
	}
	if err := a.store.TouchLogin(r.Context(), u.ID); err != nil {
		logger(r.Context()).Warn("touch login", "err", err)
	}
	a.audit(r, u.ClinicID, "auth.login", "user", u.ID.String(), map[string]any{"mfa": mfa})
	writeJSON(w, http.StatusOK, resp)
}

// loginMFA finishes a login with the TOTP code: {mfa_token, code}.
func (a *API) loginMFA(w http.ResponseWriter, r *http.Request) {
	var in struct {
		MFAToken string `json:"mfa_token"`
		Code     string `json:"code"`
	}
	if !decode(w, r, &in) {
		return
	}
	p, err := a.tokens.Verify(in.MFAToken, "mfa")
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "invalid_token", "login step expired; sign in again")
		return
	}
	u, err := a.store.GetUser(r.Context(), p.UserID)
	if errors.Is(err, domain.ErrNotFound) {
		writeErr(w, r, http.StatusUnauthorized, "invalid_token", "login step expired; sign in again")
		return
	}
	if err != nil {
		fail(w, r, err)
		return
	}
	if !u.IsActive || !u.TOTPEnabled {
		writeErr(w, r, http.StatusUnauthorized, "invalid_token", "login step expired; sign in again")
		return
	}
	if locked(u) {
		writeLocked(w, r, *u.LockedUntil)
		return
	}
	ok, err := a.checkTOTP(r.Context(), u, in.Code)
	if err != nil {
		fail(w, r, err)
		return
	}
	if !ok {
		if a.recordFailure(w, r, u, "totp") {
			return
		}
		writeErr(w, r, http.StatusUnauthorized, "invalid_code", "wrong or already used code")
		return
	}
	a.completeLogin(w, r, u, true)
}

// checkTOTP verifies a code against the user's enabled secret and burns its time step.
func (a *API) checkTOTP(ctx context.Context, u *domain.User, code string) (bool, error) {
	if len(u.TOTPSecretEnc) == 0 {
		return false, nil
	}
	secret, err := a.sealer.Open(u.TOTPSecretEnc)
	if err != nil {
		return false, fmt.Errorf("decrypt totp secret: %w", err)
	}
	step, ok := auth.VerifyTOTP(string(secret), code, time.Now())
	if !ok || step <= u.TOTPLastStep {
		return false, nil
	}
	if err := a.store.ConsumeTOTPStep(ctx, u.ID, step); errors.Is(err, domain.ErrConflict) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func emailDomain(e string) string {
	_, d, _ := strings.Cut(e, "@")
	return d
}

func (a *API) refresh(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !decode(w, r, &in) {
		return
	}
	var resp *tokenResponse
	err := a.store.InTx(r.Context(), func(tx *store.Store) error {
		uid, mfa, err := tx.ConsumeRefreshToken(r.Context(), auth.HashToken(in.RefreshToken))
		if err != nil {
			return err
		}
		u, err := tx.GetUser(r.Context(), uid)
		if err != nil {
			return err
		}
		if !u.IsActive {
			return domain.ErrUnauthorized
		}
		resp, err = a.issueTokens(r, tx, u, mfa && u.TOTPEnabled)
		return err
	})
	if errors.Is(err, store.ErrRefreshRace) {
		writeErr(w, r, http.StatusUnauthorized, "refresh_race", "refresh token was just rotated by a parallel request")
		return
	}
	if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrUnauthorized) {
		writeErr(w, r, http.StatusUnauthorized, "invalid_token", "refresh token is invalid or expired")
		return
	}
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !decode(w, r, &in) {
		return
	}
	if err := a.store.RevokeRefreshToken(r.Context(), auth.HashToken(in.RefreshToken)); err != nil {
		fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// logoutAll ends every session of the caller on every device.
func (a *API) logoutAll(w http.ResponseWriter, r *http.Request) {
	p, _ := principal(r.Context())
	if err := a.store.RevokeUserTokens(r.Context(), p.UserID); err != nil {
		fail(w, r, err)
		return
	}
	a.audit(r, p.ClinicID, "auth.logout_all", "user", p.UserID.String(), nil)
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) me(w http.ResponseWriter, r *http.Request) {
	p, _ := principal(r.Context())
	u, err := a.store.GetUser(r.Context(), p.UserID)
	if err != nil {
		fail(w, r, err)
		return
	}
	out := map[string]any{
		"user": u,
		"session": map[string]any{
			"mfa":                  p.MFA,
			"mfa_setup_required":   a.mfaGateClosed(p),
			"must_change_password": u.MustChangePassword,
		},
	}
	if u.ClinicID != nil {
		if c, err := a.store.GetClinic(r.Context(), *u.ClinicID); err == nil {
			out["clinic"] = c
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// changePassword sets a new password, ends all other sessions and returns a fresh token pair.
func (a *API) changePassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if !decode(w, r, &in) {
		return
	}
	p, _ := principal(r.Context())
	u, err := a.store.GetUser(r.Context(), p.UserID)
	if err != nil {
		fail(w, r, err)
		return
	}
	if locked(u) {
		writeLocked(w, r, *u.LockedUntil)
		return
	}
	if !auth.CheckPassword(u.PasswordHash, in.CurrentPassword) {
		if a.recordFailure(w, r, u, "password_change") {
			return
		}
		writeErr(w, r, http.StatusUnauthorized, "invalid_credentials", "current password is wrong")
		return
	}
	if in.NewPassword == in.CurrentPassword {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "new password must differ from the current one")
		return
	}
	hash, err := auth.HashPassword(in.NewPassword)
	if err != nil {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", err.Error())
		return
	}
	var resp *tokenResponse
	err = a.store.InTx(r.Context(), func(tx *store.Store) error {
		if err := tx.SetUserPassword(r.Context(), u.ID, hash, false); err != nil {
			return err
		}
		if err := tx.RevokeUserTokens(r.Context(), u.ID); err != nil {
			return err
		}
		u.MustChangePassword = false
		resp, err = a.issueTokens(r, tx, u, p.MFA)
		return err
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	a.audit(r, u.ClinicID, "auth.password_changed", "user", u.ID.String(), nil)
	writeJSON(w, http.StatusOK, resp)
}

// totpSetup creates a pending secret and returns it with the otpauth:// URL for the QR code.
// 2FA is not active until totpEnable confirms a code from the app.
func (a *API) totpSetup(w http.ResponseWriter, r *http.Request) {
	p, _ := principal(r.Context())
	u, err := a.store.GetUser(r.Context(), p.UserID)
	if err != nil {
		fail(w, r, err)
		return
	}
	if u.TOTPEnabled {
		writeErr(w, r, http.StatusConflict, "conflict", "two-factor authentication is already on")
		return
	}
	secret, err := auth.NewTOTPSecret()
	if err != nil {
		fail(w, r, err)
		return
	}
	sealed, err := a.sealer.Seal([]byte(secret))
	if err != nil {
		fail(w, r, err)
		return
	}
	if err := a.store.SetPendingTOTP(r.Context(), u.ID, sealed); err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"secret":      secret,
		"otpauth_url": auth.TOTPURL(totpIssuer, u.Email, secret),
	})
}

// totpEnable confirms the first code, turns 2FA on, ends other sessions and returns MFA tokens.
func (a *API) totpEnable(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Code string `json:"code"`
	}
	if !decode(w, r, &in) {
		return
	}
	p, _ := principal(r.Context())
	u, err := a.store.GetUser(r.Context(), p.UserID)
	if err != nil {
		fail(w, r, err)
		return
	}
	if u.TOTPEnabled {
		writeErr(w, r, http.StatusConflict, "conflict", "two-factor authentication is already on")
		return
	}
	if len(u.TOTPSecretEnc) == 0 {
		writeErr(w, r, http.StatusConflict, "conflict", "start the setup first")
		return
	}
	secret, err := a.sealer.Open(u.TOTPSecretEnc)
	if err != nil {
		fail(w, r, fmt.Errorf("decrypt totp secret: %w", err))
		return
	}
	step, ok := auth.VerifyTOTP(string(secret), in.Code, time.Now())
	if !ok {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid_code", "the code does not match; check the time on your phone")
		return
	}
	var resp *tokenResponse
	err = a.store.InTx(r.Context(), func(tx *store.Store) error {
		if err := tx.EnableTOTP(r.Context(), u.ID, step); err != nil {
			return err
		}
		if err := tx.RevokeUserTokens(r.Context(), u.ID); err != nil {
			return err
		}
		u.TOTPEnabled = true
		resp, err = a.issueTokens(r, tx, u, true)
		return err
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	a.audit(r, u.ClinicID, "auth.totp_enabled", "user", u.ID.String(), nil)
	writeJSON(w, http.StatusOK, resp)
}

// totpDisable turns 2FA off; it needs both the password and a current code.
func (a *API) totpDisable(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Password string `json:"password"`
		Code     string `json:"code"`
	}
	if !decode(w, r, &in) {
		return
	}
	p, _ := principal(r.Context())
	u, err := a.store.GetUser(r.Context(), p.UserID)
	if err != nil {
		fail(w, r, err)
		return
	}
	if !u.TOTPEnabled {
		writeErr(w, r, http.StatusConflict, "conflict", "two-factor authentication is off")
		return
	}
	if u.Role == domain.RoleSuperadmin && a.cfg.RequireSuperadminMFA {
		writeErr(w, r, http.StatusForbidden, "forbidden", "two-factor authentication is mandatory for Jazyl accounts")
		return
	}
	if locked(u) {
		writeLocked(w, r, *u.LockedUntil)
		return
	}
	if !auth.CheckPassword(u.PasswordHash, in.Password) {
		if a.recordFailure(w, r, u, "totp_disable_password") {
			return
		}
		writeErr(w, r, http.StatusUnauthorized, "invalid_credentials", "password is wrong")
		return
	}
	ok, err := a.checkTOTP(r.Context(), u, in.Code)
	if err != nil {
		fail(w, r, err)
		return
	}
	if !ok {
		if a.recordFailure(w, r, u, "totp_disable_code") {
			return
		}
		writeErr(w, r, http.StatusUnauthorized, "invalid_code", "wrong or already used code")
		return
	}
	var resp *tokenResponse
	err = a.store.InTx(r.Context(), func(tx *store.Store) error {
		if err := tx.DisableTOTP(r.Context(), u.ID); err != nil {
			return err
		}
		if err := tx.RevokeUserTokens(r.Context(), u.ID); err != nil {
			return err
		}
		u.TOTPEnabled = false
		resp, err = a.issueTokens(r, tx, u, false)
		return err
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	a.audit(r, u.ClinicID, "auth.totp_disabled", "user", u.ID.String(), nil)
	writeJSON(w, http.StatusOK, resp)
}

// wsTicket issues a 60-second token for opening the websocket (browsers cannot set headers on WS).
func (a *API) wsTicket(w http.ResponseWriter, r *http.Request) {
	p, _ := principal(r.Context())
	t, err := a.tokens.Issue(p, "ws", 60*time.Second)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ticket": t, "expires_in": 60})
}

// websocket streams live events: /api/v1/ws?ticket=...&clinic_id=... (superadmin may omit clinic_id).
func (a *API) websocket(w http.ResponseWriter, r *http.Request) {
	p, err := a.tokens.Verify(r.URL.Query().Get("ticket"), "ws")
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "invalid websocket ticket")
		return
	}
	room := r.URL.Query().Get("clinic_id")
	switch {
	case p.Role == domain.RoleSuperadmin && room == "":
		room = realtime.RoomAll()
	case p.Role == domain.RoleSuperadmin:
	case p.ClinicID != nil && (room == "" || room == p.ClinicID.String()):
		room = p.ClinicID.String()
	default:
		writeErr(w, r, http.StatusForbidden, "forbidden", "no access to this clinic")
		return
	}
	a.hub.Serve(w, r, room, wsOriginPatterns(a.cfg.DashboardOrigins))
}

func wsOriginPatterns(origins []string) []string {
	out := make([]string, 0, len(origins))
	for _, o := range origins {
		o = strings.TrimPrefix(strings.TrimPrefix(o, "https://"), "http://")
		out = append(out, strings.TrimRight(o, "/"))
	}
	return out
}

func (a *API) audit(r *http.Request, clinicID *uuid.UUID, action, entity, entityID string, meta map[string]any) {
	var uid *uuid.UUID
	if p, ok := principal(r.Context()); ok {
		uid = &p.UserID
	}
	if err := a.store.Audit(r.Context(), clinicID, uid, action, entity, entityID, a.clientIP(r), meta); err != nil {
		logger(r.Context()).Warn("audit write failed", "action", action, "err", err)
	}
}
