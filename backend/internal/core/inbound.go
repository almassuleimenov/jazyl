package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"

	"github.com/almasik/jazyl/internal/domain"
	"github.com/almasik/jazyl/internal/messaging"
	"github.com/almasik/jazyl/internal/notebook"
	"github.com/almasik/jazyl/internal/observability"
	"github.com/almasik/jazyl/internal/phone"
	"github.com/almasik/jazyl/internal/store"
	"github.com/almasik/jazyl/internal/whatsapp"
)

// Intent is a classified patient reply.
type Intent string

// Patient intents. The bot is a strict "1/2" machine, not a chatbot.
const (
	IntentConfirm    Intent = "confirm"
	IntentReschedule Intent = "reschedule"
	IntentOptOut     Intent = "opt_out"
	IntentOptIn      Intent = "opt_in"
	IntentUnknown    Intent = "unknown"
)

func normalizeText(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) {
			return -1
		}
		return r
	}, s)
	return strings.Join(strings.Fields(s), " ")
}

// ClassifyReply maps a patient's reply to an intent.
func ClassifyReply(text string) Intent {
	t := normalizeText(text)
	switch t {
	case "стоп", "stop", "отписаться", "отписка", "тоқтат", "токтат":
		return IntentOptOut
	case "старт", "start", "подписаться":
		return IntentOptIn
	}
	if strings.ContainsAny(text, "👍✅") {
		return IntentConfirm
	}
	if t == "" {
		return IntentUnknown
	}
	first := strings.Fields(t)[0]
	switch first {
	case "1", "1️⃣", "да", "приду", "придём", "придем", "буду", "ок", "ok", "окей", "конечно", "иә", "ия", "келемін", "келем", "барамын":
		return IntentConfirm
	case "2", "2️⃣", "перенести", "перенос", "нет", "не", "отмена", "отменить", "жоқ", "жок", "ауыстыру":
		return IntentReschedule
	}
	if strings.Contains(t, "не смогу") || strings.Contains(t, "не приду") || strings.Contains(t, "перенес") {
		return IntentReschedule
	}
	return IntentUnknown
}

// ParseRating extracts a 1–5 rating ("5", "5 спасибо", "⭐⭐⭐⭐"). The first number in the
// text must be a single digit 1–5 ("10" or "7" are rejected).
func ParseRating(text string) (int, bool) {
	runes := []rune(text)
	for i, r := range runes {
		if r < '0' || r > '9' {
			continue
		}
		if i+1 < len(runes) && runes[i+1] >= '0' && runes[i+1] <= '9' {
			return 0, false
		}
		if r >= '1' && r <= '5' {
			return int(r - '0'), true
		}
		return 0, false
	}
	if n := strings.Count(text, "⭐"); n >= 1 && n <= 5 {
		return n, true
	}
	return 0, false
}

// HandleWebhook processes one OpenWA delivery exactly once. A returned error makes OpenWA retry.
func (s *Service) HandleWebhook(ctx context.Context, env whatsapp.WebhookEnvelope) error {
	key := env.IdempotencyKey
	if key == "" {
		key = env.DeliveryID
	}
	err := s.store.InTx(ctx, func(tx *store.Store) error {
		if key != "" {
			first, err := tx.MarkWebhookSeen(ctx, key)
			if err != nil {
				return err
			}
			if !first {
				observability.WebhookEvents.WithLabelValues(env.Event, "duplicate").Inc()
				return nil
			}
		}
		switch env.Event {
		case "message.received":
			return s.handleIncoming(ctx, tx, env)
		case "session.status", "session.authenticated", "session.disconnected", "session.qr":
			return s.handleSessionEvent(ctx, tx, env)
		default:
			observability.WebhookEvents.WithLabelValues(env.Event, "ignored").Inc()
			return nil
		}
	})
	if err != nil {
		observability.WebhookEvents.WithLabelValues(env.Event, "error").Inc()
	}
	return err
}

func (s *Service) handleSessionEvent(ctx context.Context, tx *store.Store, env whatsapp.WebhookEnvelope) error {
	var d whatsapp.SessionStatusData
	if len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, &d); err != nil {
			return fmt.Errorf("%w: session data", domain.ErrInvalid)
		}
	}
	status := d.Status
	switch env.Event {
	case "session.authenticated":
		status = "authenticating"
	case "session.disconnected":
		status = "disconnected"
	case "session.qr":
		status = "qr_ready"
	}
	if status == "" {
		return nil
	}
	if err := tx.SetSessionStatus(ctx, env.SessionID, status); err != nil {
		return err
	}
	c, err := tx.GetClinicBySession(ctx, env.SessionID)
	if err == nil {
		s.pub.Publish(c.ID, "whatsapp.status", map[string]string{"status": status})
		if status == "disconnected" || status == "failed" || status == "action_required" {
			s.log.Warn("whatsapp session needs attention", "clinic_id", c.ID, "status", status)
		}
	}
	observability.WebhookEvents.WithLabelValues(env.Event, "ok").Inc()
	return nil
}

func (s *Service) handleIncoming(ctx context.Context, tx *store.Store, env whatsapp.WebhookEnvelope) error {
	var m whatsapp.IncomingMessage
	if err := json.Unmarshal(env.Data, &m); err != nil {
		return fmt.Errorf("%w: message data", domain.ErrInvalid)
	}
	if m.FromMe || m.IsGroup || strings.TrimSpace(m.Body) == "" {
		observability.WebhookEvents.WithLabelValues(env.Event, "ignored").Inc()
		return nil
	}
	from, ok := phone.FromChatID(m.From)
	if !ok {
		from, ok = phone.FromChatID(m.ChatID)
	}
	if !ok {
		// e.g. "@lid" privacy addresses; we cannot map them to a patient.
		s.log.Info("ignoring message from unmappable address", "from", m.From)
		observability.WebhookEvents.WithLabelValues(env.Event, "unmappable").Inc()
		return nil
	}
	c, err := tx.GetClinicBySession(ctx, env.SessionID)
	if errors.Is(err, domain.ErrNotFound) {
		s.log.Warn("webhook for unknown session", "session_id", env.SessionID)
		return nil
	}
	if err != nil {
		return err
	}

	// Paused/churned clinic: keep the message for the history, but the bot stays silent.
	if !c.Operating() {
		kind := domain.MsgPatientIn
		if c.IsAdminPhone(from) {
			kind = domain.MsgAdminIn
		}
		_, err := tx.InsertInbound(ctx, c.ID, nil, from, kind, m.Body, m.ID)
		if err == nil {
			observability.WebhookEvents.WithLabelValues(env.Event, "clinic_paused").Inc()
		}
		return err
	}

	if c.IsAdminPhone(from) {
		err = s.handleAdminMessage(ctx, tx, c, from, m)
	} else {
		err = s.handlePatientMessage(ctx, tx, c, from, m)
	}
	if err == nil {
		observability.WebhookEvents.WithLabelValues(env.Event, "ok").Inc()
	}
	return err
}

// ---------- patients ----------

func (s *Service) handlePatientMessage(ctx context.Context, tx *store.Store, c *domain.Clinic, from string, m whatsapp.IncomingMessage) error {
	p, err := tx.GetPatientByPhone(ctx, c.ID, from)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	var pid *uuid.UUID
	if p != nil {
		pid = &p.ID
	}
	in, err := tx.InsertInbound(ctx, c.ID, pid, from, domain.MsgPatientIn, m.Body, m.ID)
	if err != nil {
		return err
	}
	s.pub.Publish(c.ID, "message.received", in)

	if p == nil {
		// Unknown number writing to the clinic: the administrator answers it on their phone as usual.
		return nil
	}

	intent := ClassifyReply(m.Body)
	switch intent {
	case IntentOptOut:
		observability.BotReplies.WithLabelValues(string(intent)).Inc()
		if err := tx.SetPatientOptOut(ctx, c.ID, from, true); err != nil {
			return err
		}
		if err := tx.ClearConversation(ctx, c.ID, from); err != nil {
			return err
		}
		_, err := s.toPatient(ctx, tx, c, p, nil, domain.MsgAutoReply, messaging.OptOutAck(langOf(c, p)))
		return err
	case IntentOptIn:
		observability.BotReplies.WithLabelValues(string(intent)).Inc()
		if err := tx.SetPatientOptOut(ctx, c.ID, from, false); err != nil {
			return err
		}
		_, err := s.toPatient(ctx, tx, c, p, nil, domain.MsgAutoReply, messaging.OptInAck(langOf(c, p)))
		return err
	}

	conv, err := tx.GetConversation(ctx, c.ID, from)
	if errors.Is(err, domain.ErrNotFound) {
		return s.replyWithoutConversation(ctx, tx, c, p, intent)
	}
	if err != nil {
		return err
	}

	switch conv.Awaiting {
	case domain.AwaitReminderReply:
		if conv.AppointmentID == nil {
			return tx.ClearConversation(ctx, c.ID, from)
		}
		if intent == IntentUnknown {
			// A very short unclear reply ("?", "ок?") gets the 1/2 hint; longer texts are real
			// questions that the clinic answers itself on its own phone.
			observability.BotReplies.WithLabelValues(string(intent)).Inc()
			return s.nudgeUnknown(ctx, tx, c, p, *conv.AppointmentID, m.Body)
		}
		return s.applyReminderReply(ctx, tx, c, p, *conv.AppointmentID, intent)
	case domain.AwaitReviewRating:
		return s.applyRating(ctx, tx, c, p, conv.AppointmentID, m.Body)
	case domain.AwaitReviewComment:
		return s.applyComplaint(ctx, tx, c, p, conv.AppointmentID, m.Body)
	}
	return nil
}

// replyWithoutConversation handles a late "1"/"2" after the conversation expired.
func (s *Service) replyWithoutConversation(ctx context.Context, tx *store.Store, c *domain.Clinic, p *domain.Patient, intent Intent) error {
	if intent != IntentConfirm && intent != IntentReschedule {
		return nil
	}
	now := s.now()
	appts, err := tx.ListAppointments(ctx, c.ID, store.ApptFilter{From: now, To: now.Add(48 * time.Hour), PatientID: &p.ID})
	if err != nil {
		return err
	}
	for _, a := range appts {
		if a.Status == domain.ApptScheduled || a.Status == domain.ApptConfirmed {
			return s.applyReminderReply(ctx, tx, c, p, a.ID, intent)
		}
	}
	return nil
}

func (s *Service) applyReminderReply(ctx context.Context, tx *store.Store, c *domain.Clinic, p *domain.Patient, apptID uuid.UUID, intent Intent) error {
	lang := langOf(c, p)
	a, err := tx.GetAppointment(ctx, c.ID, apptID)
	if errors.Is(err, domain.ErrNotFound) {
		return tx.ClearConversation(ctx, c.ID, p.Phone)
	}
	if err != nil {
		return err
	}
	observability.BotReplies.WithLabelValues(string(intent)).Inc()

	switch intent {
	case IntentConfirm:
		a2, err := s.changeStatusTx(ctx, tx, c, a.ID, domain.ApptConfirmed, "patient")
		if err != nil {
			return err
		}
		if err := tx.ClearConversation(ctx, c.ID, p.Phone); err != nil {
			return err
		}
		if _, err := s.toPatient(ctx, tx, c, p, &a.ID, domain.MsgAutoReply, messaging.ConfirmAck(visitFor(c, p, a2))); err != nil {
			return err
		}
		s.pub.Publish(c.ID, "appointment.updated", a2)
		return nil

	case IntentReschedule:
		a2, err := s.changeStatusTx(ctx, tx, c, a.ID, domain.ApptRescheduleRequested, "patient")
		if err != nil {
			return err
		}
		if err := tx.ClearConversation(ctx, c.ID, p.Phone); err != nil {
			return err
		}
		if _, err := s.toPatient(ctx, tx, c, p, nil, domain.MsgAutoReply, messaging.RescheduleAck(lang)); err != nil {
			return err
		}
		if err := s.toAdmins(ctx, tx, c, messaging.AdminReschedule(p.Name, p.Phone, a.StartsAt.In(c.Location()))); err != nil {
			return err
		}
		s.pub.Publish(c.ID, "appointment.updated", a2)
		return nil
	}

	return nil
}

func (s *Service) nudgeUnknown(ctx context.Context, tx *store.Store, c *domain.Clinic, p *domain.Patient, apptID uuid.UUID, body string) error {
	if len([]rune(normalizeText(body))) > 12 {
		return nil
	}
	// Extend the conversation slightly so a follow-up "1" still lands.
	if err := tx.SetConversation(ctx, c.ID, p.Phone, domain.AwaitReminderReply, &apptID, 12*time.Hour); err != nil {
		return err
	}
	_, err := s.toPatient(ctx, tx, c, p, &apptID, domain.MsgAutoReply, messaging.UnknownReply(langOf(c, p)))
	return err
}

func (s *Service) applyRating(ctx context.Context, tx *store.Store, c *domain.Clinic, p *domain.Patient, apptID *uuid.UUID, body string) error {
	rating, ok := ParseRating(body)
	if !ok {
		return nil // free text: the clinic handles it manually
	}
	lang := langOf(c, p)
	routed := "2gis"
	if rating <= 3 {
		routed = "owner"
	}
	pid := p.ID
	if _, err := tx.CreateReview(ctx, &domain.Review{ClinicID: c.ID, AppointmentID: apptID, PatientID: &pid, Rating: rating, RoutedTo: routed}); err != nil {
		return err
	}
	observability.BotReplies.WithLabelValues("rating_" + routed).Inc()

	if routed == "2gis" {
		if err := tx.ClearConversation(ctx, c.ID, p.Phone); err != nil {
			return err
		}
		_, err := s.toPatient(ctx, tx, c, p, apptID, domain.MsgReviewLink, messaging.ReviewGood(lang, c.GisReviewURL))
		return err
	}

	visit := s.now()
	if apptID != nil {
		if a, err := tx.GetAppointment(ctx, c.ID, *apptID); err == nil {
			visit = a.StartsAt
		}
	}
	if _, err := s.toOwner(ctx, tx, c, domain.MsgOwnerNotice, messaging.OwnerComplaint(p.Name, p.Phone, rating, visit.In(c.Location()), "")); err != nil {
		return err
	}
	if err := tx.SetConversation(ctx, c.ID, p.Phone, domain.AwaitReviewComment, apptID, 24*time.Hour); err != nil {
		return err
	}
	_, err := s.toPatient(ctx, tx, c, p, apptID, domain.MsgAutoReply, messaging.ReviewBad(lang))
	s.pub.Publish(c.ID, "review.complaint", map[string]any{"patient_id": p.ID, "rating": rating})
	return err
}

func (s *Service) applyComplaint(ctx context.Context, tx *store.Store, c *domain.Clinic, p *domain.Patient, apptID *uuid.UUID, body string) error {
	if apptID != nil {
		if err := tx.SetReviewComment(ctx, c.ID, *apptID, body); err != nil {
			return err
		}
	}
	if err := tx.ClearConversation(ctx, c.ID, p.Phone); err != nil {
		return err
	}
	if _, err := s.toOwner(ctx, tx, c, domain.MsgOwnerNotice,
		fmt.Sprintf("💬 Комментарий пациента %s (+%s):\n%s", p.Name, p.Phone, strings.TrimSpace(body))); err != nil {
		return err
	}
	_, err := s.toPatient(ctx, tx, c, p, apptID, domain.MsgAutoReply, messaging.ReviewBadThanks(langOf(c, p)))
	return err
}

func langOf(c *domain.Clinic, p *domain.Patient) string {
	if p != nil && p.Language != "" {
		return p.Language
	}
	return c.Language
}

// ---------- staff (administrator / owner) ----------

func (s *Service) handleAdminMessage(ctx context.Context, tx *store.Store, c *domain.Clinic, from string, m whatsapp.IncomingMessage) error {
	if _, err := tx.InsertInbound(ctx, c.ID, nil, from, domain.MsgAdminIn, m.Body, m.ID); err != nil {
		return err
	}
	reply := func(body string) error {
		_, err := tx.EnqueueMessage(ctx, store.OutMessage{ClinicID: c.ID, Phone: from, Kind: domain.MsgAdminNotice, Body: body})
		return err
	}

	if notebook.LooksLikeSchedule(m.Body) {
		text, err := s.importScheduleTx(ctx, tx, c, from, m.Body)
		if err != nil {
			return err
		}
		return reply(text)
	}

	switch t := normalizeText(m.Body); t {
	case "помощь", "help", "көмек", "меню", "команды":
		return reply(messaging.AdminHelp())
	case "отчет", "отчёт", "есеп", "report":
		now := s.now()
		met, err := tx.ClinicMetrics(ctx, c.ID, now.AddDate(0, 0, -7), now, c.AvgCheckKZT)
		if err != nil {
			return err
		}
		return reply(RenderReport(c, "week", met))
	case "завтра", "сегодня", "ертең", "бүгін":
		day := 1
		if t == "сегодня" || t == "бүгін" {
			day = 0
		}
		text, err := s.dayAgenda(ctx, tx, c, day)
		if err != nil {
			return err
		}
		return reply(text)
	}
	if strings.ContainsAny(m.Body, "0123456789") && strings.Contains(m.Body, ":") {
		// Looks like an attempted schedule without phones: explain the format.
		res := notebook.Parse(m.Body, s.now(), c.Location())
		if len(res.Failures) > 0 {
			return reply(messaging.ScheduleAck(res.Date, nil, failureTexts(res.Failures), c.Settings.ReminderEveningHour))
		}
	}
	return nil
}

// ImportSchedule imports a notebook-style schedule (also exposed via the dashboard API).
func (s *Service) ImportSchedule(ctx context.Context, clinicID uuid.UUID, fromPhone, text string) (string, error) {
	var ack string
	err := s.store.InTx(ctx, func(tx *store.Store) error {
		c, err := tx.GetClinic(ctx, clinicID)
		if err != nil {
			return err
		}
		ack, err = s.importScheduleTx(ctx, tx, c, fromPhone, text)
		return err
	})
	if err == nil {
		s.pub.Publish(clinicID, "schedule.imported", map[string]string{"summary": ack})
	}
	return ack, err
}

func (s *Service) importScheduleTx(ctx context.Context, tx *store.Store, c *domain.Clinic, from, text string) (string, error) {
	res := notebook.Parse(text, s.now(), c.Location())
	lines := make([]messaging.ScheduleLine, 0, len(res.Entries))
	failures := failureTexts(res.Failures)
	for _, e := range res.Entries {
		if !e.StartsAt.After(s.now()) {
			failures = append(failures, fmt.Sprintf("строка %d: время %s уже прошло", e.Line, messaging.FormatTime(e.StartsAt)))
			continue
		}
		a, created, err := s.createAppointmentTx(ctx, tx, c, NewAppointment{
			PatientPhone: e.Phone, PatientName: e.Name, ServiceName: e.Service, StartsAt: e.StartsAt, Source: "notebook",
		})
		if err != nil {
			if errors.Is(err, domain.ErrInvalid) {
				failures = append(failures, fmt.Sprintf("строка %d: %v", e.Line, err))
				continue
			}
			return "", err
		}
		lines = append(lines, messaging.ScheduleLine{At: a.StartsAt.In(c.Location()), Name: a.PatientName, New: created})
	}
	if err := tx.InsertScheduleImport(ctx, c.ID, from, text, res.Date, len(lines), res.Failures); err != nil {
		return "", err
	}
	return messaging.ScheduleAck(res.Date, lines, failures, c.Settings.ReminderEveningHour), nil
}

func failureTexts(fs []notebook.Failure) []string {
	out := make([]string, 0, len(fs))
	for _, f := range fs {
		out = append(out, fmt.Sprintf("строка %d «%s» — %s", f.Line, f.Text, f.Reason))
	}
	return out
}

func (s *Service) dayAgenda(ctx context.Context, tx *store.Store, c *domain.Clinic, offset int) (string, error) {
	loc := c.Location()
	n := s.now().In(loc)
	start := time.Date(n.Year(), n.Month(), n.Day()+offset, 0, 0, 0, 0, loc)
	appts, err := tx.ListAppointments(ctx, c.ID, store.ApptFilter{From: start, To: start.AddDate(0, 0, 1)})
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "📅 %s — записей: %d\n", messaging.FormatDate(start, "ru"), len(appts))
	icon := map[string]string{
		domain.ApptScheduled: "⏳", domain.ApptConfirmed: "✅", domain.ApptRescheduleRequested: "🔁",
		domain.ApptCancelled: "❌", domain.ApptCompleted: "✔️", domain.ApptNoShow: "🚫",
	}
	for _, a := range appts {
		fmt.Fprintf(&b, "%s %s %s\n", icon[a.Status], messaging.FormatTime(a.StartsAt.In(loc)), a.PatientName)
	}
	b.WriteString("\n✅ подтвердил · ⏳ ждём ответа · 🔁 просит перенести")
	return b.String(), nil
}
