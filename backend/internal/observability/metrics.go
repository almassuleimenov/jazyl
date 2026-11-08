// Package observability holds Prometheus metrics shared across the app.
package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Registry is the app's Prometheus registry (exposed at /metrics).
var Registry = prometheus.NewRegistry()

var (
	// HTTPRequests counts HTTP requests by route pattern, method and status class.
	HTTPRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "jazyl_http_requests_total", Help: "HTTP requests.",
	}, []string{"route", "method", "code"})

	// HTTPDuration observes request latency.
	HTTPDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "jazyl_http_request_duration_seconds", Help: "HTTP request latency.",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
	}, []string{"route", "method"})

	// MessagesSent counts outbound WhatsApp messages by kind and result (sent|retry|failed).
	MessagesSent = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "jazyl_messages_total", Help: "Outbound WhatsApp message attempts.",
	}, []string{"kind", "result"})

	// SendLatency observes provider send latency.
	SendLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "jazyl_whatsapp_send_seconds", Help: "WhatsApp provider send latency.",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 9),
	})

	// RemindersProcessed counts reminders by kind and outcome (sent|skipped).
	RemindersProcessed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "jazyl_reminders_total", Help: "Processed reminders.",
	}, []string{"kind", "outcome"})

	// WebhookEvents counts inbound webhook events by event type and outcome.
	WebhookEvents = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "jazyl_webhook_events_total", Help: "Inbound OpenWA webhook events.",
	}, []string{"event", "outcome"})

	// BotReplies counts classified patient replies.
	BotReplies = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "jazyl_bot_replies_total", Help: "Patient replies handled by the bot.",
	}, []string{"intent"})

	// Leads counts incoming leads by source.
	Leads = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "jazyl_leads_total", Help: "Incoming leads.",
	}, []string{"source"})

	// OutboxQueued is the current depth of the outbox (sampled by the worker).
	OutboxQueued = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "jazyl_outbox_queued", Help: "Queued outbound messages.",
	})

	// WSConnections is the number of live dashboard websocket connections.
	WSConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "jazyl_ws_connections", Help: "Open websocket connections.",
	})

	// JobRuns counts background job ticks by job and result.
	JobRuns = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "jazyl_job_runs_total", Help: "Background job runs.",
	}, []string{"job", "result"})
)

func init() {
	Registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		HTTPRequests, HTTPDuration, MessagesSent, SendLatency, RemindersProcessed,
		WebhookEvents, BotReplies, Leads, OutboxQueued, WSConnections, JobRuns,
	)
}
