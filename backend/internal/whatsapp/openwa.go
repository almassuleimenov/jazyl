package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/almasik/jazyl/internal/phone"
)

// OpenWA is a client for the OpenWA gateway REST API (https://github.com/rmyndharis/OpenWA).
// Base URL example: http://openwa:2785 (the client appends /api).
type OpenWA struct {
	base   string
	apiKey string
	http   *http.Client
}

// NewOpenWA creates a client.
func NewOpenWA(baseURL, apiKey string) *OpenWA {
	return &OpenWA{
		base:   baseURL + "/api",
		apiKey: apiKey,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

// APIError is a non-2xx OpenWA response.
type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string { return fmt.Sprintf("openwa: HTTP %d: %s", e.Status, e.Body) }

func (c *OpenWA) do(ctx context.Context, method, path string, in, out any, headers map[string]string) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("openwa %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("read openwa response: %w", err)
	}

	if resp.StatusCode >= 300 {
		apiErr := &APIError{Status: resp.StatusCode, Body: string(truncateBytes(raw, 500))}
		switch {
		case resp.StatusCode == http.StatusConflict:
			return fmt.Errorf("%w: %w", ErrNotReady, apiErr)
		case resp.StatusCode == http.StatusTooManyRequests:
			return &RateLimitedError{RetryAfter: retryAfter(resp, raw)}
		case resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity ||
			resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden:
			return fmt.Errorf("%w: %w", ErrPermanent, apiErr)
		default:
			return apiErr // 401, 5xx: retryable
		}
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("decode openwa response: %w", err)
		}
	}
	return nil
}

func retryAfter(resp *http.Response, raw []byte) time.Duration {
	if s, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && s > 0 {
		return time.Duration(s) * time.Second
	}
	var body struct {
		RetryAfterSeconds int `json:"retryAfterSeconds"`
	}
	if json.Unmarshal(raw, &body) == nil && body.RetryAfterSeconds > 0 {
		return time.Duration(body.RetryAfterSeconds) * time.Second
	}
	return time.Minute
}

func truncateBytes(b []byte, n int) []byte {
	if len(b) > n {
		return b[:n]
	}
	return b
}

// SendText implements Sender via POST /sessions/{id}/messages/send-text.
func (c *OpenWA) SendText(ctx context.Context, sessionID, to, text, idemKey string) (string, error) {
	if sessionID == "" {
		return "", ErrNoSession
	}
	var out struct {
		MessageID string `json:"messageId"`
	}
	headers := map[string]string{}
	if idemKey != "" {
		headers["Idempotency-Key"] = idemKey
	}
	err := c.do(ctx, http.MethodPost, "/sessions/"+url.PathEscape(sessionID)+"/messages/send-text",
		map[string]any{"chatId": phone.ChatID(to), "text": text}, &out, headers)
	if err != nil {
		return "", err
	}
	return out.MessageID, nil
}

// CreateSession implements SessionManager.
func (c *OpenWA) CreateSession(ctx context.Context, name string) (*SessionInfo, error) {
	var out SessionInfo
	if err := c.do(ctx, http.MethodPost, "/sessions", map[string]any{"name": name}, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// StartSession implements SessionManager.
func (c *OpenWA) StartSession(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodPost, "/sessions/"+url.PathEscape(id)+"/start", nil, nil, nil)
}

// GetSession implements SessionManager.
func (c *OpenWA) GetSession(ctx context.Context, id string) (*SessionInfo, error) {
	var out SessionInfo
	if err := c.do(ctx, http.MethodGet, "/sessions/"+url.PathEscape(id), nil, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// QRCode implements SessionManager.
func (c *OpenWA) QRCode(ctx context.Context, id string) (string, string, error) {
	var out struct {
		QRCode string `json:"qrCode"`
		Status string `json:"status"`
	}
	if err := c.do(ctx, http.MethodGet, "/sessions/"+url.PathEscape(id)+"/qr", nil, &out, nil); err != nil {
		return "", "", err
	}
	return out.QRCode, out.Status, nil
}

// PairingCode implements SessionManager (link by 8-char code instead of QR).
func (c *OpenWA) PairingCode(ctx context.Context, id, phoneNumber string) (string, error) {
	var out struct {
		PairingCode string `json:"pairingCode"`
	}
	if err := c.do(ctx, http.MethodPost, "/sessions/"+url.PathEscape(id)+"/pairing-code",
		map[string]any{"phoneNumber": phoneNumber}, &out, nil); err != nil {
		return "", err
	}
	return out.PairingCode, nil
}

// RegisterWebhook implements SessionManager.
func (c *OpenWA) RegisterWebhook(ctx context.Context, sessionID, hookURL, secret string, events []string) error {
	return c.do(ctx, http.MethodPost, "/sessions/"+url.PathEscape(sessionID)+"/webhooks", map[string]any{
		"url":        hookURL,
		"events":     events,
		"secret":     secret,
		"retryCount": 5,
	}, nil, nil)
}

// DeleteSession implements SessionManager: unlinks the phone and drops the session in OpenWA.
func (c *OpenWA) DeleteSession(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/sessions/"+url.PathEscape(id), nil, nil, nil)
}

// Health pings the gateway.
func (c *OpenWA) Health(ctx context.Context) error {
	err := c.do(ctx, http.MethodGet, "/health", nil, nil, nil)
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.Status == http.StatusUnauthorized {
		return nil // reachable
	}
	return err
}

// WebhookEnvelope is the OpenWA webhook body.
type WebhookEnvelope struct {
	Event          string          `json:"event"`
	Timestamp      string          `json:"timestamp"`
	SessionID      string          `json:"sessionId"`
	IdempotencyKey string          `json:"idempotencyKey"`
	DeliveryID     string          `json:"deliveryId"`
	Data           json.RawMessage `json:"data"`
}

// IncomingMessage is data of a message.received event.
type IncomingMessage struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	ChatID    string `json:"chatId"`
	Body      string `json:"body"`
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"`
	IsGroup   bool   `json:"isGroup"`
	FromMe    bool   `json:"fromMe"`
}

// SessionStatusData is data of a session.* event.
type SessionStatusData struct {
	Status string `json:"status"`
	Phone  string `json:"phone"`
}
