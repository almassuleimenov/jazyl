package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/almasik/jazyl/internal/auth"
	"github.com/almasik/jazyl/internal/domain"
	"github.com/almasik/jazyl/internal/realtime"
	"github.com/almasik/jazyl/internal/store"
)

type tokenResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	User         *domain.User `json:"user"`
}

func (a *API) issueTokens(w http.ResponseWriter, r *http.Request, tx *store.Store, u *domain.User) (*tokenResponse, error) {
	p := auth.Principal{UserID: u.ID, Role: u.Role, ClinicID: u.ClinicID}
	access, err := a.tokens.Issue(p, "", a.cfg.AccessTokenTTL)
	if err != nil {
		return nil, err
	}
	refresh, hash, err := auth.RandomToken("jzr_", 32)
	if err != nil {
		return nil, err
	}
	if err := tx.InsertRefreshToken(r.Context(), u.ID, hash, time.Now().Add(a.cfg.RefreshTokenTTL), r.UserAgent(), a.clientIP(r)); err != nil {
		return nil, err
	}
	return &tokenResponse{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer", ExpiresIn: int(a.cfg.AccessTokenTTL.Seconds()), User: u}, nil
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
	hash := ""
	if u != nil {
		hash = u.PasswordHash
	}
	if !auth.CheckPassword(hash, in.Password) || u == nil || !u.IsActive {
		logger(r.Context()).Warn("login failed", "email_domain", emailDomain(in.Email), "ip", a.clientIP(r))
		writeErr(w, r, http.StatusUnauthorized, "invalid_credentials", "wrong email or password")
		return
	}
	resp, err := a.issueTokens(w, r, a.store, u)
	if err != nil {
		fail(w, r, err)
		return
	}
	if err := a.store.TouchLogin(r.Context(), u.ID); err != nil {
		logger(r.Context()).Warn("touch login", "err", err)
	}
	a.audit(r, u.ClinicID, "auth.login", "user", u.ID.String(), nil)
	writeJSON(w, http.StatusOK, resp)
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
		uid, err := tx.ConsumeRefreshToken(r.Context(), auth.HashToken(in.RefreshToken))
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
		resp, err = a.issueTokens(w, r, tx, u)
		return err
	})
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

func (a *API) me(w http.ResponseWriter, r *http.Request) {
	p, _ := principal(r.Context())
	u, err := a.store.GetUser(r.Context(), p.UserID)
	if err != nil {
		fail(w, r, err)
		return
	}
	out := map[string]any{"user": u}
	if u.ClinicID != nil {
		if c, err := a.store.GetClinic(r.Context(), *u.ClinicID); err == nil {
			out["clinic"] = c
		}
	}
	writeJSON(w, http.StatusOK, out)
}

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
	if !auth.CheckPassword(u.PasswordHash, in.CurrentPassword) {
		writeErr(w, r, http.StatusUnauthorized, "invalid_credentials", "current password is wrong")
		return
	}
	hash, err := auth.HashPassword(in.NewPassword)
	if err != nil {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", err.Error())
		return
	}
	err = a.store.InTx(r.Context(), func(tx *store.Store) error {
		if err := tx.SetUserPassword(r.Context(), u.ID, hash); err != nil {
			return err
		}
		return tx.RevokeUserTokens(r.Context(), u.ID)
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	a.audit(r, u.ClinicID, "auth.password_changed", "user", u.ID.String(), nil)
	w.WriteHeader(http.StatusNoContent)
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
