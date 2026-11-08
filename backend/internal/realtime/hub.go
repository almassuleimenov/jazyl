// Package realtime pushes live events (new lead, patient confirmed, session status...) to the
// dashboard over WebSocket. Each connection is subscribed to exactly one clinic room
// (superadmins may subscribe to "*" to receive every clinic's events).
package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/almasik/jazyl/internal/observability"
)

// Event is pushed to clients.
type Event struct {
	Type     string    `json:"type"`
	ClinicID uuid.UUID `json:"clinic_id"`
	At       time.Time `json:"at"`
	Data     any       `json:"data,omitempty"`
}

// Publisher publishes events (implemented by Hub; a no-op is fine in tests).
type Publisher interface {
	Publish(clinicID uuid.UUID, typ string, data any)
}

// Nop discards events.
type Nop struct{}

// Publish implements Publisher.
func (Nop) Publish(uuid.UUID, string, any) {}

const allRoom = "*"

type client struct {
	send chan []byte
}

// Hub fan-outs events to subscribed clients.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*client]struct{}
	log   *slog.Logger
}

// NewHub creates an empty hub.
func NewHub(log *slog.Logger) *Hub {
	return &Hub{rooms: map[string]map[*client]struct{}{}, log: log.With("component", "ws")}
}

// Publish implements Publisher. Slow clients drop events rather than blocking producers.
func (h *Hub) Publish(clinicID uuid.UUID, typ string, data any) {
	b, err := json.Marshal(Event{Type: typ, ClinicID: clinicID, At: time.Now().UTC(), Data: data})
	if err != nil {
		h.log.Error("marshal event", "err", err)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, room := range []string{clinicID.String(), allRoom} {
		for c := range h.rooms[room] {
			select {
			case c.send <- b:
			default: // buffer full: client is too slow, skip
			}
		}
	}
}

func (h *Hub) add(room string, c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[room] == nil {
		h.rooms[room] = map[*client]struct{}{}
	}
	h.rooms[room][c] = struct{}{}
	observability.WSConnections.Inc()
}

func (h *Hub) remove(room string, c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms[room], c)
	if len(h.rooms[room]) == 0 {
		delete(h.rooms, room)
	}
	observability.WSConnections.Dec()
}

// Serve upgrades the request and streams events of room (a clinic id or "*") until disconnect.
// originPatterns restricts browser origins (e.g. "app.jazyl.kz", "localhost:3000").
func (h *Hub) Serve(w http.ResponseWriter, r *http.Request, room string, originPatterns []string) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: originPatterns})
	if err != nil {
		h.log.Debug("ws accept failed", "err", err)
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(4096)

	c := &client{send: make(chan []byte, 64)}
	h.add(room, c)
	defer h.remove(room, c)

	// The reader only handles control frames and detects client close.
	ctx := conn.CloseRead(r.Context())

	ping := time.NewTicker(25 * time.Second)
	defer ping.Stop()

	hello, _ := json.Marshal(Event{Type: "hello", At: time.Now().UTC(), Data: map[string]string{"room": room}}) // static value, cannot fail
	if err := writeTimeout(ctx, conn, hello); err != nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.send:
			if err := writeTimeout(ctx, conn, msg); err != nil {
				return
			}
		case <-ping.C:
			pctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := conn.Ping(pctx)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

func writeTimeout(ctx context.Context, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return c.Write(ctx, websocket.MessageText, msg)
}

// RoomAll is the superadmin room.
func RoomAll() string { return allRoom }
