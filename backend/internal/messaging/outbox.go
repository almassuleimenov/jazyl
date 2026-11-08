package messaging

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"

	"github.com/almasik/jazyl/internal/domain"
	"github.com/almasik/jazyl/internal/observability"
	"github.com/almasik/jazyl/internal/phone"
	"github.com/almasik/jazyl/internal/store"
	"github.com/almasik/jazyl/internal/whatsapp"
)

const (
	maxAttempts   = 8
	claimBatch    = 20
	claimLease    = 2 * time.Minute
	pollInterval  = 2 * time.Second
	expireAfter   = 24 * time.Hour
	reminderStale = 6 * time.Hour
)

// Outbox delivers queued messages with per-number pacing (anti-ban) and retries.
type Outbox struct {
	store       *store.Store
	router      *whatsapp.Router
	log         *slog.Logger
	minInterval time.Duration
	jitter      time.Duration

	nextAllowed map[string]time.Time // sender key -> earliest next send (owned by Run's goroutine)
	now         func() time.Time
}

// NewOutbox creates the worker.
func NewOutbox(st *store.Store, router *whatsapp.Router, log *slog.Logger, minInterval, jitter time.Duration) *Outbox {
	return &Outbox{
		store: st, router: router, log: log.With("component", "outbox"),
		minInterval: minInterval, jitter: jitter,
		nextAllowed: map[string]time.Time{},
		now:         time.Now,
	}
}

// Run processes the outbox until ctx is cancelled.
func (o *Outbox) Run(ctx context.Context) error {
	tick := time.NewTicker(pollInterval)
	defer tick.Stop()
	housekeeping := time.NewTicker(time.Minute)
	defer housekeeping.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-housekeeping.C:
			if n, err := o.store.ReleaseStuckMessages(ctx); err != nil {
				o.log.Error("release stuck messages", "err", err)
			} else if n > 0 {
				o.log.Warn("re-queued stuck messages", "count", n)
			}
			o.sampleDepth(ctx)
		case <-tick.C:
			if err := o.ProcessBatch(ctx); err != nil && ctx.Err() == nil {
				o.log.Error("outbox batch", "err", err)
				observability.JobRuns.WithLabelValues("outbox", "error").Inc()
			}
		}
	}
}

func (o *Outbox) sampleDepth(ctx context.Context) {
	var n int
	if err := o.store.Pool().QueryRow(ctx, `SELECT count(*) FROM messages WHERE status='queued'`).Scan(&n); err == nil {
		observability.OutboxQueued.Set(float64(n))
	}
}

// ProcessBatch claims and sends one batch. Exported for tests and manual triggers.
func (o *Outbox) ProcessBatch(ctx context.Context) error {
	items, err := o.store.ClaimOutbox(ctx, claimBatch, claimLease)
	if err != nil {
		return err
	}
	for _, it := range items {
		if ctx.Err() != nil {
			// Shutting down: put the remaining claimed rows back immediately.
			o.requeue(context.WithoutCancel(ctx), it.Message.ID, o.now(), "shutdown", true)
			continue
		}
		o.deliver(ctx, it)
	}
	return nil
}

func senderKey(it store.OutboxItem) string {
	if it.SessionID != nil {
		return it.Provider + ":" + *it.SessionID
	}
	return it.Provider + ":" + it.Message.ClinicID.String()
}

func (o *Outbox) deliver(ctx context.Context, it store.OutboxItem) {
	m := it.Message
	now := o.now()
	log := o.log.With("message_id", m.ID, "clinic_id", m.ClinicID, "kind", m.Kind, "to", phone.Mask(m.Phone))

	// Drop messages that are no longer useful.
	if age := now.Sub(m.CreatedAt); age > expireAfter || (m.Kind == domain.MsgReminder && age > reminderStale) {
		o.fail(ctx, m, "expired before delivery")
		return
	}

	key := senderKey(it)
	if next, ok := o.nextAllowed[key]; ok && now.Before(next) {
		o.requeue(ctx, m.ID, next, "paced", true)
		return
	}

	sender, err := o.router.Get(it.Provider)
	if err != nil {
		o.fail(ctx, m, err.Error())
		return
	}
	session := ""
	if it.SessionID != nil {
		session = *it.SessionID
	}

	sendCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	start := time.Now()
	providerID, err := sender.SendText(sendCtx, session, m.Phone, m.Body, "jazyl-"+m.ID.String())
	cancel()
	observability.SendLatency.Observe(time.Since(start).Seconds())

	// Pace the number regardless of outcome: failed attempts also count as activity.
	o.nextAllowed[key] = o.now().Add(o.minInterval + o.randJitter())

	if err == nil {
		if err := o.store.MarkMessageSent(ctx, m.ID, providerID); err != nil {
			log.Error("mark sent", "err", err)
		}
		observability.MessagesSent.WithLabelValues(m.Kind, "sent").Inc()
		log.Info("message sent")
		return
	}

	var rl *whatsapp.RateLimitedError
	switch {
	case errors.As(err, &rl):
		o.nextAllowed[key] = o.now().Add(rl.RetryAfter)
		o.requeue(ctx, m.ID, o.now().Add(rl.RetryAfter), err.Error(), true)
	case errors.Is(err, whatsapp.ErrPermanent):
		o.fail(ctx, m, err.Error())
	case errors.Is(err, whatsapp.ErrNoSession), errors.Is(err, whatsapp.ErrNotReady):
		// The clinic's phone is not linked/online. Keep the message until it expires.
		delete(o.nextAllowed, key)
		o.requeue(ctx, m.ID, o.now().Add(3*time.Minute), err.Error(), true)
		log.Warn("whatsapp session not ready", "err", err)
	default:
		if m.Attempts >= maxAttempts {
			o.fail(ctx, m, err.Error())
			return
		}
		o.requeue(ctx, m.ID, o.now().Add(Backoff(m.Attempts)), err.Error(), false)
		log.Warn("send failed, will retry", "attempt", m.Attempts, "err", err)
	}
}

func (o *Outbox) requeue(ctx context.Context, id uuid.UUID, at time.Time, reason string, refund bool) {
	if err := o.store.RetryMessage(ctx, id, at, reason, refund); err != nil {
		o.log.Error("requeue message", "message_id", id, "err", err)
	}
	if reason != "paced" && reason != "shutdown" {
		observability.MessagesSent.WithLabelValues("any", "retry").Inc()
	}
}

func (o *Outbox) fail(ctx context.Context, m *domain.Message, reason string) {
	if err := o.store.FailMessage(ctx, m.ID, reason); err != nil {
		o.log.Error("fail message", "message_id", m.ID, "err", err)
	}
	observability.MessagesSent.WithLabelValues(m.Kind, "failed").Inc()
	o.log.Warn("message failed", "message_id", m.ID, "kind", m.Kind, "reason", reason)
}

func (o *Outbox) randJitter() time.Duration {
	if o.jitter <= 0 {
		return 0
	}
	return time.Duration(rand.Int64N(int64(o.jitter)))
}

// Backoff returns exponential backoff for attempt n (1-based): 30s, 1m, 2m ... capped at 30m.
func Backoff(n int) time.Duration {
	if n < 1 {
		n = 1
	}
	d := 30 * time.Second
	for i := 1; i < n && d < 30*time.Minute; i++ {
		d *= 2
	}
	if d > 30*time.Minute {
		d = 30 * time.Minute
	}
	return d
}
