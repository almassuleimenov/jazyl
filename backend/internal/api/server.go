package api

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/almasik/jazyl/internal/auth"
	"github.com/almasik/jazyl/internal/config"
	"github.com/almasik/jazyl/internal/core"
	"github.com/almasik/jazyl/internal/domain"
	"github.com/almasik/jazyl/internal/observability"
	"github.com/almasik/jazyl/internal/realtime"
	"github.com/almasik/jazyl/internal/store"
	"github.com/almasik/jazyl/internal/whatsapp"
)

// API holds handler dependencies.
type API struct {
	cfg      *config.Config
	log      *slog.Logger
	store    *store.Store
	svc      *core.Service
	tokens   *auth.Tokens
	sealer   *auth.Sealer
	hub      *realtime.Hub
	sessions whatsapp.SessionManager
	router   *whatsapp.Router

	loginLimiter   *Limiter // login + 2FA step, per client IP
	refreshLimiter *Limiter // token refresh, per client IP
	accountLimiter *Limiter // password / 2FA changes, per user
	publicLimiter  *Limiter
	siteLimiter    *Limiter
}

// Deps are the API's collaborators.
type Deps struct {
	Config   *config.Config
	Log      *slog.Logger
	Store    *store.Store
	Service  *core.Service
	Tokens   *auth.Tokens
	Hub      *realtime.Hub
	Sessions whatsapp.SessionManager
	Router   *whatsapp.Router
}

// New builds the HTTP handler.
func New(d Deps) (http.Handler, error) {
	sealer, err := auth.NewSealer(d.Config.DataEncryptionKey)
	if err != nil {
		return nil, err
	}
	a := &API{
		sealer: sealer,
		cfg:    d.Config, log: d.Log, store: d.Store, svc: d.Service, tokens: d.Tokens, hub: d.Hub,
		sessions: d.Sessions, router: d.Router,
		loginLimiter:   NewLimiter(d.Config.LoginRatePerMinute, 10),
		refreshLimiter: NewLimiter(120, 30),
		accountLimiter: NewLimiter(10, 5),
		publicLimiter:  NewLimiter(20, 10),
		siteLimiter:    NewLimiter(300, 60),
	}
	return a.routes(), nil
}

func (a *API) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(a.requestContext, a.recoverer, securityHeaders)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/readyz", a.readyz)
	r.Handle("/metrics", a.metricsAuth(promhttp.HandlerFor(observability.Registry, promhttp.HandlerOpts{})))

	// OpenWA → Jazyl webhook (HMAC-signed).
	r.Post("/webhooks/openwa", a.openwaWebhook)

	// Public API for clinic websites (hosted separately, e.g. Vercel).
	r.Route("/api/public/v1", func(r chi.Router) {
		r.Use(a.siteCORS)
		r.Use(a.rateLimit(a.publicLimiter, a.clientIP))
		r.Use(a.requireSiteKey)
		r.Get("/clinic", a.publicClinic)
		r.Get("/slots", a.publicSlots)
		r.Post("/bookings", a.publicBooking)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(a.dashboardCORS)

		r.With(a.rateLimit(a.loginLimiter, a.clientIP)).Post("/auth/login", a.login)
		r.With(a.rateLimit(a.loginLimiter, a.clientIP)).Post("/auth/mfa", a.loginMFA)
		r.With(a.rateLimit(a.refreshLimiter, a.clientIP)).Post("/auth/refresh", a.refresh)
		r.Post("/auth/logout", a.logout)
		r.Get("/ws", a.websocket) // authenticated by a short-lived ticket

		// Account security: reachable while a password change or 2FA setup is still pending.
		r.Group(func(r chi.Router) {
			r.Use(a.requireAuth)
			r.Get("/auth/me", a.me)
			r.With(a.rateLimit(a.accountLimiter, a.principalKey)).Post("/auth/password", a.changePassword)
			r.Post("/auth/logout-all", a.logoutAll)
			r.Post("/auth/totp/setup", a.totpSetup)
			r.With(a.rateLimit(a.accountLimiter, a.principalKey)).Post("/auth/totp/enable", a.totpEnable)
			r.With(a.rateLimit(a.accountLimiter, a.principalKey)).Post("/auth/totp/disable", a.totpDisable)
		})

		// Everything else needs a session that changed its temporary password and,
		// for Jazyl staff, passed 2FA.
		r.Group(func(r chi.Router) {
			r.Use(a.requireAuth, a.requireSessionReady)
			r.Post("/ws/ticket", a.wsTicket)

			// Platform (Jazyl team) endpoints.
			r.Route("/admin", func(r chi.Router) {
				r.Use(requireRole(domain.RoleSuperadmin))
				r.Get("/overview", a.platformOverview)
				r.Get("/clinics", a.listClinics)
				r.Post("/clinics", a.createClinic)
				r.Delete("/clinics/{clinicID}", a.deleteClinic)
				r.Post("/clinics/{clinicID}/whatsapp/connect", a.whatsappConnect)
				r.Get("/clinics/{clinicID}/whatsapp", a.whatsappStatus)
				r.Get("/clinics/{clinicID}/whatsapp/qr", a.whatsappQR)
				r.Post("/clinics/{clinicID}/whatsapp/pairing-code", a.whatsappPairing)
			})

			// Clinic-scoped endpoints (tenant isolation enforced by requireClinicAccess).
			r.Route("/clinics/{clinicID}", func(r chi.Router) {
				r.Use(requireClinicAccess)
				r.Get("/", a.getClinic)
				r.With(requireRole(domain.RoleSuperadmin, domain.RoleOwner)).Patch("/", a.updateClinic)

				r.With(requireRole(domain.RoleSuperadmin, domain.RoleOwner)).Get("/users", a.listUsers)
				r.With(requireRole(domain.RoleSuperadmin, domain.RoleOwner)).Post("/users", a.createUser)
				r.With(requireRole(domain.RoleSuperadmin, domain.RoleOwner)).Patch("/users/{userID}", a.updateUser)
				r.With(requireRole(domain.RoleSuperadmin, domain.RoleOwner)).Post("/users/{userID}/password", a.resetUserPassword)
				r.With(requireRole(domain.RoleSuperadmin)).Post("/users/{userID}/reset-2fa", a.resetUserMFA)

				r.Get("/doctors", a.listDoctors)
				r.Post("/doctors", a.upsertDoctor)
				r.Put("/doctors/{id}", a.upsertDoctor)
				r.Get("/services", a.listServices)
				r.Post("/services", a.upsertService)
				r.Put("/services/{id}", a.upsertService)

				r.Get("/patients", a.listPatients)
				r.Get("/patients/{id}", a.getPatient)
				r.Patch("/patients/{id}", a.updatePatient)

				r.Get("/appointments", a.listAppointments)
				r.Post("/appointments", a.createAppointment)
				r.Get("/appointments/{id}", a.getAppointment)
				r.Patch("/appointments/{id}", a.updateAppointment)
				r.Post("/appointments/{id}/status", a.setAppointmentStatus)
				r.Post("/schedule/import", a.importSchedule)

				r.Get("/leads", a.listLeads)
				r.Patch("/leads/{id}", a.updateLead)
				r.Post("/leads/{id}/convert", a.convertLead)

				r.Get("/messages", a.listMessages)
				r.Post("/messages", a.sendManualMessage)
				r.Get("/chats", a.listChats)
				r.Get("/reviews", a.listReviews)

				r.Get("/analytics/summary", a.analyticsSummary)
				r.Get("/analytics/daily", a.analyticsDaily)
				r.Get("/reports", a.listReports)
				r.Get("/reports/preview", a.previewReport)
				r.With(requireRole(domain.RoleSuperadmin, domain.RoleOwner)).Post("/reports/send", a.sendReport)

				r.With(requireRole(domain.RoleSuperadmin, domain.RoleOwner)).Get("/site-keys", a.listSiteKeys)
				r.With(requireRole(domain.RoleSuperadmin, domain.RoleOwner)).Post("/site-keys", a.createSiteKey)
				r.With(requireRole(domain.RoleSuperadmin, domain.RoleOwner)).Delete("/site-keys/{id}", a.revokeSiteKey)
			})
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeErr(w, r, http.StatusNotFound, "not_found", "route not found")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeErr(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	})
	return r
}

func (a *API) readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	checks := map[string]string{"database": "ok"}
	status := http.StatusOK
	if err := a.store.Pool().Ping(ctx); err != nil {
		checks["database"] = "down"
		status = http.StatusServiceUnavailable
	}
	if h, ok := a.sessions.(interface{ Health(context.Context) error }); ok {
		if err := h.Health(ctx); err != nil {
			checks["openwa"] = "down" // degraded: messages wait in the outbox, API still serves
		} else {
			checks["openwa"] = "ok"
		}
	}
	writeJSON(w, status, checks)
}

func (a *API) metricsAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.cfg.MetricsToken != "" {
			got := r.Header.Get("Authorization")
			want := "Bearer " + a.cfg.MetricsToken
			if subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
				writeErr(w, r, http.StatusUnauthorized, "unauthorized", "metrics token required")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
