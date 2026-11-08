// Package whatsapp abstracts the WhatsApp transport. Today it is OpenWA (self-hosted gateway);
// an official Cloud API provider can be added behind the same Sender interface later.
package whatsapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/almasik/jazyl/internal/phone"
)

// Errors classify send failures so the outbox can decide whether to retry.
var (
	// ErrNotReady: session is not connected (QR not scanned, phone offline). Retry later.
	ErrNotReady = errors.New("whatsapp session not ready")
	// ErrPermanent: the request can never succeed (bad number, validation error).
	ErrPermanent = errors.New("permanent whatsapp error")
	// ErrNoSession: the clinic has no WhatsApp session configured.
	ErrNoSession = errors.New("clinic has no whatsapp session")
)

// RateLimitedError asks the caller to wait.
type RateLimitedError struct{ RetryAfter time.Duration }

func (e *RateLimitedError) Error() string {
	return fmt.Sprintf("rate limited, retry after %s", e.RetryAfter)
}

// Sender sends a text message.
type Sender interface {
	// SendText sends text to a digits-only phone from sessionID. idemKey de-duplicates retries.
	SendText(ctx context.Context, sessionID, to, text, idemKey string) (providerMessageID string, err error)
}

// SessionInfo is the state of a WhatsApp session.
type SessionInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Phone  string `json:"phone,omitempty"`
}

// SessionManager links clinics' WhatsApp numbers.
type SessionManager interface {
	CreateSession(ctx context.Context, name string) (*SessionInfo, error)
	StartSession(ctx context.Context, id string) error
	GetSession(ctx context.Context, id string) (*SessionInfo, error)
	QRCode(ctx context.Context, id string) (qrDataURL, status string, err error)
	PairingCode(ctx context.Context, id, phone string) (code string, err error)
	RegisterWebhook(ctx context.Context, sessionID, url, secret string, events []string) error
}

// Mock is an in-memory provider for local development and tests. It "delivers" by logging.
type Mock struct {
	Log *slog.Logger
	mu  sync.Mutex
	// Sent keeps the last messages for tests.
	Sent []MockMessage
}

// MockMessage is a message captured by Mock.
type MockMessage struct {
	SessionID, To, Text string
}

// SendText implements Sender.
func (m *Mock) SendText(_ context.Context, sessionID, to, text, _ string) (string, error) {
	m.mu.Lock()
	m.Sent = append(m.Sent, MockMessage{SessionID: sessionID, To: to, Text: text})
	if len(m.Sent) > 1000 {
		m.Sent = m.Sent[len(m.Sent)-1000:]
	}
	m.mu.Unlock()
	if m.Log != nil {
		m.Log.Info("mock whatsapp send", "session", sessionID, "to", phone.Mask(to), "chars", len([]rune(text)))
		m.Log.Debug("mock whatsapp text", "text", text)
	}
	return "mock_" + uuid.NewString(), nil
}

// Router picks a Sender per clinic provider name.
type Router struct {
	providers map[string]Sender
}

// NewRouter creates a provider router.
func NewRouter(providers map[string]Sender) *Router { return &Router{providers: providers} }

// Get returns the sender for a provider.
func (r *Router) Get(provider string) (Sender, error) {
	s, ok := r.providers[provider]
	if !ok {
		return nil, fmt.Errorf("%w: unknown provider %q", ErrPermanent, provider)
	}
	return s, nil
}

// VerifySignature checks an OpenWA webhook signature: "sha256=" + hex(HMAC-SHA256(secret, rawBody)).
func VerifySignature(body []byte, header, secret string) bool {
	if secret == "" || header == "" {
		return false
	}
	got, ok := strings.CutPrefix(strings.TrimSpace(header), "sha256=")
	if !ok {
		return false
	}
	sig, err := hex.DecodeString(got)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hmac.Equal(sig, mac.Sum(nil))
}

// Sign produces a signature header value (used by tests and the mock webhook tool).
func Sign(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
