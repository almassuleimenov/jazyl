// Package api is the HTTP layer: routing, middleware, handlers.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/almasik/jazyl/internal/core"
	"github.com/almasik/jazyl/internal/domain"
	"github.com/almasik/jazyl/internal/whatsapp"
)

// apiError is the error envelope: {"error": {"code": "...", "message": "...", "request_id": "..."}}.
type apiError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Default().Debug("write response", "err", err)
	}
}

func writeErr(w http.ResponseWriter, r *http.Request, status int, code, msg string) {
	writeJSON(w, status, map[string]apiError{"error": {Code: code, Message: msg, RequestID: requestID(r.Context())}})
}

// fail maps domain errors to HTTP responses; unexpected errors are logged and hidden.
func fail(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeErr(w, r, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, domain.ErrConflict):
		writeErr(w, r, http.StatusConflict, "conflict", publicMsg(err))
	case errors.Is(err, domain.ErrInvalid):
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", publicMsg(err))
	case errors.Is(err, domain.ErrUnauthorized):
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "authentication required")
	case errors.Is(err, domain.ErrForbidden):
		writeErr(w, r, http.StatusForbidden, "forbidden", "access denied")
	case errors.Is(err, core.ErrPlanRequired):
		writeErr(w, r, http.StatusPaymentRequired, "plan_required", "this feature requires the «Полное кресло» plan or higher")
	case errors.Is(err, whatsapp.ErrPermanent), errors.Is(err, whatsapp.ErrNotReady):
		// Log the gateway's reason: without it a 502 in the logs is undiagnosable.
		logger(r.Context()).Warn("whatsapp gateway error", "err", err)
		writeErr(w, r, http.StatusBadGateway, "whatsapp_error", err.Error())
	default:
		logger(r.Context()).Error("request failed", "err", err)
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal error")
	}
}

func publicMsg(err error) string {
	// Domain errors are wrapped as "<sentinel>: detail"; expose the full chain (no internals inside).
	return err.Error()
}

const maxBody = 1 << 20

// decode reads a JSON body strictly (unknown fields rejected).
func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var mbe *http.MaxBytesError
		switch {
		case errors.As(err, &mbe):
			writeErr(w, r, http.StatusRequestEntityTooLarge, "too_large", "request body too large")
		case errors.Is(err, io.EOF):
			writeErr(w, r, http.StatusBadRequest, "bad_request", "request body is empty")
		default:
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid JSON: "+err.Error())
		}
		return false
	}
	if dec.More() {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "body must contain a single JSON object")
		return false
	}
	return true
}

func pathUUID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid "+name)
		return uuid.Nil, false
	}
	return id, true
}

func queryUUID(r *http.Request, name string) (*uuid.UUID, error) {
	v := r.URL.Query().Get(name)
	if v == "" {
		return nil, nil
	}
	id, err := uuid.Parse(v)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrInvalid, name)
	}
	return &id, nil
}

func queryInt(r *http.Request, name string, def int) int {
	n, err := strconv.Atoi(r.URL.Query().Get(name))
	if err != nil {
		return def
	}
	return n
}

// queryRange parses from/to (RFC3339 or YYYY-MM-DD in loc). Defaults: last `days` days.
func queryRange(r *http.Request, loc *time.Location, days int) (time.Time, time.Time, error) {
	parse := func(name string) (time.Time, error) {
		v := strings.TrimSpace(r.URL.Query().Get(name))
		if v == "" {
			return time.Time{}, nil
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t, nil
		}
		t, err := time.ParseInLocation("2006-01-02", v, loc)
		if err != nil {
			return time.Time{}, fmt.Errorf("%w: %s must be RFC3339 or YYYY-MM-DD", domain.ErrInvalid, name)
		}
		return t, nil
	}
	from, err := parse("from")
	if err != nil {
		return from, from, err
	}
	to, err := parse("to")
	if err != nil {
		return from, to, err
	}
	if to.IsZero() {
		n := time.Now().In(loc)
		to = time.Date(n.Year(), n.Month(), n.Day()+1, 0, 0, 0, 0, loc)
	}
	if from.IsZero() {
		from = to.AddDate(0, 0, -days)
	}
	if !to.After(from) || to.Sub(from) > 400*24*time.Hour {
		return from, to, fmt.Errorf("%w: range must be positive and at most 400 days", domain.ErrInvalid)
	}
	return from, to, nil
}

// pageResponse wraps a list with a next cursor.
type pageResponse[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func newPage[T any](items []T, limit int, cursorOf func(T) string) pageResponse[T] {
	if items == nil {
		items = []T{}
	}
	p := pageResponse[T]{Items: items}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if len(items) == limit {
		p.NextCursor = cursorOf(items[len(items)-1])
	}
	return p
}
