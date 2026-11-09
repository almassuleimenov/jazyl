package api

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/almasik/jazyl/internal/auth"
	"github.com/almasik/jazyl/internal/domain"
	"github.com/almasik/jazyl/internal/phone"
	"github.com/almasik/jazyl/internal/store"
)

var reSlug = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$`)

type clinicInput struct {
	Slug         *string          `json:"slug"`
	Name         *string          `json:"name"`
	City         *string          `json:"city"`
	Address      *string          `json:"address"`
	Timezone     *string          `json:"timezone"`
	Language     *string          `json:"language"`
	Plan         *string          `json:"plan"`
	Status       *string          `json:"status"`
	OwnerPhone   *string          `json:"owner_phone"`
	AdminPhones  *[]string        `json:"admin_phones"`
	GisReviewURL *string          `json:"gis_review_url"`
	AvgCheckKZT  *int             `json:"avg_check_kzt"`
	Settings     *domain.Settings `json:"settings"`
	WAProvider   *string          `json:"wa_provider"`
}

// apply validates and merges input into c. platform=true allows plan/status/provider changes.
func (in *clinicInput) apply(c *domain.Clinic, platform bool) error {
	set := func(dst *string, v *string) {
		if v != nil {
			*dst = strings.TrimSpace(*v)
		}
	}
	set(&c.Name, in.Name)
	set(&c.City, in.City)
	set(&c.Address, in.Address)
	set(&c.GisReviewURL, in.GisReviewURL)
	if in.Timezone != nil {
		if _, err := time.LoadLocation(*in.Timezone); err != nil {
			return fmt.Errorf("%w: timezone", domain.ErrInvalid)
		}
		c.Timezone = *in.Timezone
	}
	if in.Language != nil {
		if *in.Language != "ru" && *in.Language != "kk" {
			return fmt.Errorf("%w: language must be ru or kk", domain.ErrInvalid)
		}
		c.Language = *in.Language
	}
	if in.OwnerPhone != nil {
		if *in.OwnerPhone == "" {
			c.OwnerPhone = ""
		} else {
			p, err := phone.Normalize(*in.OwnerPhone)
			if err != nil {
				return fmt.Errorf("%w: owner_phone", domain.ErrInvalid)
			}
			c.OwnerPhone = p
		}
	}
	if in.AdminPhones != nil {
		out := make([]string, 0, len(*in.AdminPhones))
		for _, raw := range *in.AdminPhones {
			p, err := phone.Normalize(raw)
			if err != nil {
				return fmt.Errorf("%w: admin_phones %q", domain.ErrInvalid, raw)
			}
			out = append(out, p)
		}
		c.AdminPhones = out
	}
	if in.AvgCheckKZT != nil {
		if *in.AvgCheckKZT < 0 {
			return fmt.Errorf("%w: avg_check_kzt", domain.ErrInvalid)
		}
		c.AvgCheckKZT = *in.AvgCheckKZT
	}
	if in.Settings != nil {
		s := *in.Settings
		s.Normalize()
		c.Settings = s
	}
	if in.GisReviewURL != nil && c.GisReviewURL != "" && !strings.HasPrefix(c.GisReviewURL, "https://") {
		return fmt.Errorf("%w: gis_review_url must be https", domain.ErrInvalid)
	}

	if in.Plan != nil || in.Status != nil || in.WAProvider != nil {
		if !platform {
			return fmt.Errorf("%w: plan, status and wa_provider are managed by Jazyl", domain.ErrForbidden)
		}
	}
	if in.Plan != nil {
		switch *in.Plan {
		case domain.PlanNoShow, domain.PlanFullChair, domain.PlanGrowth:
			c.Plan = *in.Plan
		default:
			return fmt.Errorf("%w: plan", domain.ErrInvalid)
		}
	}
	if in.Status != nil {
		switch *in.Status {
		case "pilot", "active", "paused", "churned":
			c.Status = *in.Status
		default:
			return fmt.Errorf("%w: status", domain.ErrInvalid)
		}
	}
	if in.WAProvider != nil {
		if *in.WAProvider != "openwa" && *in.WAProvider != "mock" {
			return fmt.Errorf("%w: wa_provider", domain.ErrInvalid)
		}
		c.WAProvider = *in.WAProvider
	}
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("%w: name is required", domain.ErrInvalid)
	}
	return nil
}

func (a *API) createClinic(w http.ResponseWriter, r *http.Request) {
	var in clinicInput
	if !decode(w, r, &in) {
		return
	}
	if in.Slug == nil || !reSlug.MatchString(*in.Slug) {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "slug: 3-50 chars, lowercase latin letters, digits and dashes")
		return
	}
	c := &domain.Clinic{
		Slug: *in.Slug, Timezone: "Asia/Almaty", Language: "ru", Plan: domain.PlanNoShow, Status: "pilot",
		AdminPhones: []string{}, AvgCheckKZT: 15000, Settings: domain.DefaultSettings(), WAProvider: "openwa",
	}
	if err := in.apply(c, true); err != nil {
		fail(w, r, err)
		return
	}
	out, err := a.store.CreateClinic(r.Context(), c)
	if err != nil {
		fail(w, r, err)
		return
	}
	a.audit(r, &out.ID, "clinic.created", "clinic", out.ID.String(), map[string]any{"slug": out.Slug})
	writeJSON(w, http.StatusCreated, out)
}

func (a *API) listClinics(w http.ResponseWriter, r *http.Request) {
	cs, err := a.store.ListClinics(r.Context())
	if err != nil {
		fail(w, r, err)
		return
	}
	if cs == nil {
		cs = []*domain.Clinic{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": cs})
}

func (a *API) platformOverview(w http.ResponseWriter, r *http.Request) {
	rows, err := a.store.PlatformOverview(r.Context())
	if err != nil {
		fail(w, r, err)
		return
	}
	mrr, paying := 0, 0
	byPlan := map[string]int{}
	for _, row := range rows {
		byPlan[row.Plan]++
		if row.Status == "active" {
			mrr += domain.PlanPriceKZT(row.Plan)
			paying++
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"clinics": rows, "total": len(rows), "paying": paying, "mrr_kzt": mrr, "by_plan": byPlan,
	})
}

func (a *API) getClinic(w http.ResponseWriter, r *http.Request) {
	cid, _ := pathUUID(w, r, "clinicID")
	c, err := a.store.GetClinic(r.Context(), cid)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (a *API) updateClinic(w http.ResponseWriter, r *http.Request) {
	cid, _ := pathUUID(w, r, "clinicID")
	var in clinicInput
	if !decode(w, r, &in) {
		return
	}
	if in.Slug != nil {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "slug cannot be changed")
		return
	}
	c, err := a.store.GetClinic(r.Context(), cid)
	if err != nil {
		fail(w, r, err)
		return
	}
	p, _ := principal(r.Context())
	if err := in.apply(c, p.Role == domain.RoleSuperadmin); err != nil {
		fail(w, r, err)
		return
	}
	out, err := a.store.UpdateClinic(r.Context(), c)
	if err != nil {
		fail(w, r, err)
		return
	}
	a.audit(r, &cid, "clinic.updated", "clinic", cid.String(), nil)
	writeJSON(w, http.StatusOK, out)
}

// ---------- users ----------

func (a *API) listUsers(w http.ResponseWriter, r *http.Request) {
	cid, _ := pathUUID(w, r, "clinicID")
	us, err := a.store.ListClinicUsers(r.Context(), cid)
	if err != nil {
		fail(w, r, err)
		return
	}
	if us == nil {
		us = []*domain.User{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": us})
}

func (a *API) createUser(w http.ResponseWriter, r *http.Request) {
	cid, _ := pathUUID(w, r, "clinicID")
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Role     string `json:"role"`
	}
	if !decode(w, r, &in) {
		return
	}
	p, _ := principal(r.Context())
	if in.Role != domain.RoleOwner && in.Role != domain.RoleAdmin {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "role must be owner or admin")
		return
	}
	if in.Role == domain.RoleOwner && p.Role != domain.RoleSuperadmin {
		writeErr(w, r, http.StatusForbidden, "forbidden", "only Jazyl can create owners")
		return
	}
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if !strings.Contains(email, "@") || len(email) > 254 {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "email")
		return
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", err.Error())
		return
	}
	// The password was chosen by someone else, so the user must replace it at first login.
	u, err := a.store.CreateUser(r.Context(), &domain.User{ClinicID: &cid, Email: email, PasswordHash: hash,
		Name: strings.TrimSpace(in.Name), Role: in.Role, MustChangePassword: true})
	if errors.Is(err, domain.ErrConflict) {
		writeErr(w, r, http.StatusConflict, "conflict", "a user with this email already exists")
		return
	}
	if err != nil {
		fail(w, r, err)
		return
	}
	a.audit(r, &cid, "user.created", "user", u.ID.String(), map[string]any{"role": u.Role})
	writeJSON(w, http.StatusCreated, u)
}

func (a *API) updateUser(w http.ResponseWriter, r *http.Request) {
	cid, _ := pathUUID(w, r, "clinicID")
	uid, ok := pathUUID(w, r, "userID")
	if !ok {
		return
	}
	var in struct {
		IsActive *bool `json:"is_active"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.IsActive == nil {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "is_active is required")
		return
	}
	if p, _ := principal(r.Context()); p.UserID == uid {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "you cannot deactivate yourself")
		return
	}
	if err := a.store.SetUserActive(r.Context(), cid, uid, *in.IsActive); err != nil {
		fail(w, r, err)
		return
	}
	if !*in.IsActive {
		if err := a.store.RevokeUserTokens(r.Context(), uid); err != nil {
			fail(w, r, err)
			return
		}
	}
	a.audit(r, &cid, "user.updated", "user", uid.String(), map[string]any{"is_active": *in.IsActive})
	w.WriteHeader(http.StatusNoContent)
}

// resetUserPassword sets a temporary password for a clinic user (Jazyl, or the owner for admins).
// The user must change it at next login; all their sessions end and any lockout is cleared.
func (a *API) resetUserPassword(w http.ResponseWriter, r *http.Request) {
	cid, _ := pathUUID(w, r, "clinicID")
	uid, ok := pathUUID(w, r, "userID")
	if !ok {
		return
	}
	var in struct {
		Password string `json:"password"`
	}
	if !decode(w, r, &in) {
		return
	}
	p, _ := principal(r.Context())
	if p.UserID == uid {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "use the account page to change your own password")
		return
	}
	target, err := a.store.GetClinicUser(r.Context(), cid, uid)
	if err != nil {
		fail(w, r, err)
		return
	}
	if p.Role != domain.RoleSuperadmin && target.Role != domain.RoleAdmin {
		writeErr(w, r, http.StatusForbidden, "forbidden", "only Jazyl can reset an owner's password")
		return
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", err.Error())
		return
	}
	err = a.store.InTx(r.Context(), func(tx *store.Store) error {
		if err := tx.SetUserPassword(r.Context(), uid, hash, true); err != nil {
			return err
		}
		return tx.RevokeUserTokens(r.Context(), uid)
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	a.audit(r, &cid, "user.password_reset", "user", uid.String(), nil)
	w.WriteHeader(http.StatusNoContent)
}

// resetUserMFA removes a clinic user's 2FA (lost phone). Jazyl only.
func (a *API) resetUserMFA(w http.ResponseWriter, r *http.Request) {
	cid, _ := pathUUID(w, r, "clinicID")
	uid, ok := pathUUID(w, r, "userID")
	if !ok {
		return
	}
	if _, err := a.store.GetClinicUser(r.Context(), cid, uid); err != nil {
		fail(w, r, err)
		return
	}
	err := a.store.InTx(r.Context(), func(tx *store.Store) error {
		if err := tx.DisableTOTP(r.Context(), uid); err != nil {
			return err
		}
		return tx.RevokeUserTokens(r.Context(), uid)
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	a.audit(r, &cid, "user.totp_reset", "user", uid.String(), nil)
	w.WriteHeader(http.StatusNoContent)
}

// ---------- WhatsApp session (superadmin) ----------

func (a *API) whatsappConnect(w http.ResponseWriter, r *http.Request) {
	cid, ok := pathUUID(w, r, "clinicID")
	if !ok {
		return
	}
	c, err := a.store.GetClinic(r.Context(), cid)
	if err != nil {
		fail(w, r, err)
		return
	}
	if c.WAProvider == "mock" {
		mockID := "mock-" + c.Slug
		if err := a.store.SetClinicSession(r.Context(), cid, &mockID, "ready"); err != nil {
			fail(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"session_id": mockID, "status": "ready"})
		return
	}
	if a.cfg.OpenWAWebhookSecret == "" {
		writeErr(w, r, http.StatusFailedDependency, "not_configured", "OPENWA_WEBHOOK_SECRET is not set")
		return
	}

	sessionID := ""
	// A session stuck in "created" never got its webhook or never started (e.g. the first attempt
	// failed on a bad PUBLIC_BASE_URL); register the webhook again. This only happens before the
	// phone is linked; inbound events are also deduplicated by OpenWA's idempotency key.
	needHook := true
	if c.WASessionID != nil {
		sessionID = *c.WASessionID
		needHook = c.WASessionStatus == "" || c.WASessionStatus == "created"
	} else {
		s, err := a.sessions.CreateSession(r.Context(), "jazyl-"+c.Slug)
		if err != nil {
			fail(w, r, fmt.Errorf("create openwa session: %w", err))
			return
		}
		sessionID = s.ID
		if err := a.store.SetClinicSession(r.Context(), cid, &sessionID, s.Status); err != nil {
			fail(w, r, err)
			return
		}
	}
	if needHook {
		hook := a.cfg.PublicBaseURL + "/webhooks/openwa"
		events := []string{"message.received", "session.status", "session.authenticated", "session.disconnected"}
		if err := a.sessions.RegisterWebhook(r.Context(), sessionID, hook, a.cfg.OpenWAWebhookSecret, events); err != nil {
			fail(w, r, fmt.Errorf("register webhook %s: %w", hook, err))
			return
		}
	}
	if err := a.sessions.StartSession(r.Context(), sessionID); err != nil {
		fail(w, r, fmt.Errorf("start session: %w", err))
		return
	}
	a.audit(r, &cid, "whatsapp.connect", "clinic", cid.String(), map[string]any{"session_id": sessionID})
	writeJSON(w, http.StatusOK, map[string]string{
		"session_id": sessionID, "status": "initializing",
		"next": "GET /api/v1/admin/clinics/" + cid.String() + "/whatsapp/qr and scan it with the clinic phone",
	})
}

func (a *API) sessionOf(w http.ResponseWriter, r *http.Request) (uuid.UUID, string, bool) {
	cid, ok := pathUUID(w, r, "clinicID")
	if !ok {
		return cid, "", false
	}
	c, err := a.store.GetClinic(r.Context(), cid)
	if err != nil {
		fail(w, r, err)
		return cid, "", false
	}
	if c.WASessionID == nil {
		writeErr(w, r, http.StatusConflict, "not_connected", "call whatsapp/connect first")
		return cid, "", false
	}
	return cid, *c.WASessionID, true
}

func (a *API) whatsappStatus(w http.ResponseWriter, r *http.Request) {
	cid, sid, ok := a.sessionOf(w, r)
	if !ok {
		return
	}
	if strings.HasPrefix(sid, "mock-") {
		writeJSON(w, http.StatusOK, map[string]string{"session_id": sid, "status": "ready"})
		return
	}
	s, err := a.sessions.GetSession(r.Context(), sid)
	if err != nil {
		fail(w, r, err)
		return
	}
	if err := a.store.SetSessionStatus(r.Context(), sid, s.Status); err != nil {
		logger(r.Context()).Warn("cache session status", "err", err)
	}
	a.hub.Publish(cid, "whatsapp.status", map[string]string{"status": s.Status})
	writeJSON(w, http.StatusOK, s)
}

func (a *API) whatsappQR(w http.ResponseWriter, r *http.Request) {
	_, sid, ok := a.sessionOf(w, r)
	if !ok {
		return
	}
	qr, status, err := a.sessions.QRCode(r.Context(), sid)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"qr_code": qr, "status": status})
}

func (a *API) whatsappPairing(w http.ResponseWriter, r *http.Request) {
	_, sid, ok := a.sessionOf(w, r)
	if !ok {
		return
	}
	var in struct {
		Phone string `json:"phone"`
	}
	if !decode(w, r, &in) {
		return
	}
	p, err := phone.Normalize(in.Phone)
	if err != nil {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "phone")
		return
	}
	code, err := a.sessions.PairingCode(r.Context(), sid, p)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pairing_code": code})
}
