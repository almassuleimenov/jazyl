package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/almasik/jazyl/internal/auth"
	"github.com/almasik/jazyl/internal/core"
	"github.com/almasik/jazyl/internal/domain"
	"github.com/almasik/jazyl/internal/whatsapp"
)

// siteCORS: preflights are answered for any origin (they carry no key); the actual request's
// Origin is checked against the key's allowed_origins in requireSiteKey.
func (a *API) siteCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			h := w.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Add("Vary", "Origin")
			h.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Content-Type, X-Site-Key")
			h.Set("Access-Control-Max-Age", "600")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type siteCtx struct {
	ClinicID uuid.UUID
	KeyID    uuid.UUID
}

// requireSiteKey authenticates a clinic website by its X-Site-Key.
func (a *API) requireSiteKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimSpace(r.Header.Get("X-Site-Key"))
		if !strings.HasPrefix(key, "jzs_") {
			writeErr(w, r, http.StatusUnauthorized, "unauthorized", "missing or invalid X-Site-Key")
			return
		}
		k, err := a.store.GetActiveSiteKey(r.Context(), auth.HashToken(key))
		if errors.Is(err, domain.ErrNotFound) {
			writeErr(w, r, http.StatusUnauthorized, "unauthorized", "missing or invalid X-Site-Key")
			return
		}
		if err != nil {
			fail(w, r, err)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" && len(k.AllowedOrigins) > 0 && !contains(k.AllowedOrigins, origin) {
			writeErr(w, r, http.StatusForbidden, "origin_not_allowed", "this origin is not allowed for the site key")
			return
		}
		if !a.siteLimiter.Allow(k.ID.String()) {
			writeErr(w, r, http.StatusTooManyRequests, "rate_limited", "too many requests for this site")
			return
		}
		ctx := context.WithValue(r.Context(), keySiteClinic, siteCtx{ClinicID: k.ClinicID, KeyID: k.ID})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func siteClinic(r *http.Request) uuid.UUID {
	s, _ := r.Context().Value(keySiteClinic).(siteCtx)
	return s.ClinicID
}

// publicClinic returns what a website needs to render the booking form.
func (a *API) publicClinic(w http.ResponseWriter, r *http.Request) {
	cid := siteClinic(r)
	c, err := a.store.GetClinic(r.Context(), cid)
	if err != nil {
		fail(w, r, err)
		return
	}
	svcs, err := a.store.ListServices(r.Context(), cid, true)
	if err != nil {
		fail(w, r, err)
		return
	}
	docs, err := a.store.ListDoctors(r.Context(), cid, true)
	if err != nil {
		fail(w, r, err)
		return
	}
	type pubService struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		DurationMin int       `json:"duration_min"`
		PriceKZT    int       `json:"price_kzt"`
	}
	type pubDoctor struct {
		ID        uuid.UUID `json:"id"`
		Name      string    `json:"name"`
		Specialty string    `json:"specialty"`
	}
	ps := make([]pubService, 0, len(svcs))
	for _, s := range svcs {
		ps = append(ps, pubService{s.ID, s.Name, s.DurationMin, s.PriceKZT})
	}
	pd := make([]pubDoctor, 0, len(docs))
	for _, d := range docs {
		pd = append(pd, pubDoctor{d.ID, d.Name, d.Specialty})
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(w, http.StatusOK, map[string]any{
		"name": c.Name, "city": c.City, "address": c.Address, "timezone": c.Timezone, "language": c.Language,
		"working_hours": c.Settings.WorkingHours, "services": ps, "doctors": pd,
		"online_booking": domain.PlanAllowsSite(c.Plan) || c.Status == "pilot",
	})
}

func (a *API) publicSlots(w http.ResponseWriter, r *http.Request) {
	cid := siteClinic(r)
	c, err := a.store.GetClinic(r.Context(), cid)
	if err != nil {
		fail(w, r, err)
		return
	}
	date, err := time.ParseInLocation("2006-01-02", r.URL.Query().Get("date"), c.Location())
	if err != nil {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "date must be YYYY-MM-DD")
		return
	}
	if date.After(time.Now().AddDate(0, 3, 0)) {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "date is too far ahead")
		return
	}
	svc, err := queryUUID(r, "service_id")
	if err != nil {
		fail(w, r, err)
		return
	}
	slots, err := a.svc.FreeSlots(r.Context(), cid, date, svc)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"date": date.Format("2006-01-02"), "slots": slots})
}

func (a *API) publicBooking(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name        string            `json:"name"`
		Phone       string            `json:"phone"`
		ServiceID   *uuid.UUID        `json:"service_id"`
		Service     string            `json:"service"`
		PreferredAt *time.Time        `json:"preferred_at"`
		Comment     string            `json:"comment"`
		PageURL     string            `json:"page_url"`
		UTM         map[string]string `json:"utm"`
		Source      string            `json:"source"`
		Language    string            `json:"language"`
		Consent     bool              `json:"consent"`
		Website     string            `json:"website"` // honeypot: real users never fill it
	}
	if !decode(w, r, &in) {
		return
	}
	if in.Website != "" {
		// Pretend success to bots.
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "received"})
		return
	}
	if !in.Consent {
		writeErr(w, r, http.StatusUnprocessableEntity, "consent_required", "patient consent to personal data processing is required")
		return
	}
	if len([]rune(in.Name)) > 100 || len([]rune(in.Comment)) > 1000 || len(in.PageURL) > 500 || len([]rune(in.Service)) > 200 || len(in.UTM) > 10 {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "field too long")
		return
	}
	switch in.Source {
	case "", "site", "instagram", "2gis", "other":
	default:
		in.Source = "other"
	}
	if in.Language != "kk" {
		in.Language = "ru"
	}
	res, err := a.svc.CreateSiteBooking(r.Context(), siteClinic(r), core.SiteBooking{
		Name: in.Name, Phone: in.Phone, ServiceID: in.ServiceID, Service: in.Service, PreferredAt: in.PreferredAt,
		Comment: strings.TrimSpace(in.Comment), PageURL: in.PageURL, UTM: in.UTM, Source: in.Source, Language: in.Language,
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusAccepted, res)
}

// ---------- OpenWA webhook ----------

func (a *API) openwaWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 2<<20))
	if err != nil {
		writeErr(w, r, http.StatusRequestEntityTooLarge, "too_large", "body too large")
		return
	}
	if a.cfg.OpenWAWebhookSecret == "" {
		if a.cfg.IsProduction() {
			writeErr(w, r, http.StatusServiceUnavailable, "not_configured", "webhook secret not configured")
			return
		}
		logger(r.Context()).Warn("OPENWA_WEBHOOK_SECRET is empty: accepting unsigned webhook (dev only)")
	} else if !whatsapp.VerifySignature(body, r.Header.Get("X-OpenWA-Signature"), a.cfg.OpenWAWebhookSecret) {
		writeErr(w, r, http.StatusUnauthorized, "bad_signature", "invalid webhook signature")
		return
	}

	var env whatsapp.WebhookEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if env.IdempotencyKey == "" {
		env.IdempotencyKey = r.Header.Get("X-OpenWA-Idempotency-Key")
	}
	if env.Event == "" {
		env.Event = r.Header.Get("X-OpenWA-Event")
	}
	if err := a.svc.HandleWebhook(r.Context(), env); err != nil {
		if errors.Is(err, domain.ErrInvalid) {
			// Malformed payload: retrying will not help.
			writeErr(w, r, http.StatusUnprocessableEntity, "invalid", err.Error())
			return
		}
		fail(w, r, err) // 5xx → OpenWA retries
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
