package api

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/almasik/jazyl/internal/auth"
	"github.com/almasik/jazyl/internal/domain"
	"github.com/almasik/jazyl/internal/observability"
)

type ctxKey int

const (
	keyRequestID ctxKey = iota
	keyLogger
	keyPrincipal
	keySiteClinic
)

func requestID(ctx context.Context) string {
	s, _ := ctx.Value(keyRequestID).(string)
	return s
}

func logger(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(keyLogger).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

func principal(ctx context.Context) (auth.Principal, bool) {
	p, ok := ctx.Value(keyPrincipal).(auth.Principal)
	return p, ok
}

// statusRecorder captures the response status.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Unwrap lets http.ResponseController reach the underlying writer (websocket hijack).
func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }

// requestContext assigns a request id, a scoped logger, logs and records metrics.
func (a *API) requestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" || len(rid) > 64 {
			rid = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", rid)
		l := a.log.With("request_id", rid)
		ctx := context.WithValue(context.WithValue(r.Context(), keyRequestID, rid), keyLogger, l)

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, r.WithContext(ctx))
		dur := time.Since(start)

		route := chi.RouteContext(ctx).RoutePattern()
		if route == "" {
			route = "unmatched"
		}
		observability.HTTPRequests.WithLabelValues(route, r.Method, strconv.Itoa(rec.status/100)+"xx").Inc()
		observability.HTTPDuration.WithLabelValues(route, r.Method).Observe(dur.Seconds())
		if route != "/healthz" && route != "/metrics" {
			lvl := slog.LevelInfo
			if rec.status >= 500 {
				lvl = slog.LevelError
			}
			l.Log(ctx, lvl, "http", "method", r.Method, "route", route, "status", rec.status,
				"duration_ms", dur.Milliseconds(), "ip", a.clientIP(r))
		}
	})
}

func (a *API) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				if v == http.ErrAbortHandler {
					panic(v)
				}
				logger(r.Context()).Error("panic", "value", v, "stack", string(debug.Stack()))
				writeErr(w, r, http.StatusInternalServerError, "internal", "internal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cache-Control", "no-store")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

// dashboardCORS allows the Jazyl dashboard origins (credentials are bearer tokens, not cookies).
func (a *API) dashboardCORS(next http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, o := range a.cfg.DashboardOrigins {
		allowed[strings.TrimRight(o, "/")] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && allowed[origin] {
			h := w.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Add("Vary", "Origin")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			h.Set("Access-Control-Max-Age", "600")
		}
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireAuth validates the bearer access token.
func (a *API) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		tok, ok := strings.CutPrefix(h, "Bearer ")
		if !ok || tok == "" {
			writeErr(w, r, http.StatusUnauthorized, "unauthorized", "missing bearer token")
			return
		}
		p, err := a.tokens.Verify(tok, "")
		if err != nil {
			writeErr(w, r, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), keyPrincipal, p)
		ctx = context.WithValue(ctx, keyLogger, logger(ctx).With("user_id", p.UserID))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requireSessionReady blocks sessions that must first change a temporary password, and
// superadmin sessions without 2FA when that is required. The dashboard routes on these codes.
func (a *API) requireSessionReady(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, _ := principal(r.Context())
		if p.MustChangePassword {
			writeErr(w, r, http.StatusForbidden, "password_change_required", "change the temporary password first")
			return
		}
		if a.mfaGateClosed(p) {
			writeErr(w, r, http.StatusForbidden, "mfa_required", "two-factor authentication is required for this account")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// principalKey rate-limits per authenticated user.
func (a *API) principalKey(r *http.Request) string {
	if p, ok := principal(r.Context()); ok {
		return "u:" + p.UserID.String()
	}
	return a.clientIP(r)
}

func requireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := principal(r.Context())
			if !ok {
				writeErr(w, r, http.StatusUnauthorized, "unauthorized", "authentication required")
				return
			}
			for _, role := range roles {
				if p.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeErr(w, r, http.StatusForbidden, "forbidden", "insufficient role")
		})
	}
}

// requireClinicAccess enforces tenant isolation: superadmins see all, others only their clinic.
func requireClinicAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cid, ok := pathUUID(w, r, "clinicID")
		if !ok {
			return
		}
		p, _ := principal(r.Context())
		if p.Role != domain.RoleSuperadmin && (p.ClinicID == nil || *p.ClinicID != cid) {
			// 404 rather than 403: do not reveal that another tenant's id exists.
			writeErr(w, r, http.StatusNotFound, "not_found", "resource not found")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP returns the caller IP. X-Forwarded-For is trusted only behind a known proxy.
func (a *API) clientIP(r *http.Request) string {
	if a.cfg.TrustedProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ip, _, _ := strings.Cut(xff, ",")
			return strings.TrimSpace(ip)
		}
		if rip := r.Header.Get("X-Real-IP"); rip != "" {
			return rip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ---------- rate limiting (in-memory token bucket; per instance) ----------

type bucket struct {
	tokens float64
	last   time.Time
}

// Limiter is a keyed token bucket.
type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64 // tokens per second
	burst   float64
}

// NewLimiter allows `burst` requests immediately and refills at perMinute/60 per second.
func NewLimiter(perMinute, burst int) *Limiter {
	return &Limiter{buckets: map[string]*bucket{}, rate: float64(perMinute) / 60, burst: float64(burst)}
}

// Allow consumes a token for key.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	b, ok := l.buckets[key]
	if !ok {
		if len(l.buckets) > 50_000 { // crude memory guard
			l.buckets = map[string]*bucket{}
		}
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}
	b.tokens += now.Sub(b.last).Seconds() * l.rate
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (a *API) rateLimit(l *Limiter, keyFn func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !l.Allow(keyFn(r)) {
				w.Header().Set("Retry-After", "60")
				writeErr(w, r, http.StatusTooManyRequests, "rate_limited", "too many requests, try again later")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
