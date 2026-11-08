package core

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/almasik/jazyl/internal/domain"
	"github.com/almasik/jazyl/internal/messaging"
	"github.com/almasik/jazyl/internal/store"
)

// ReportPeriod returns [start, end) of the report that is due at now for period, in clinic time.
// Weekly: the 7 days ending today (sent Friday evening). Monthly: the previous calendar month.
func ReportPeriod(period string, now time.Time, loc *time.Location) (time.Time, time.Time) {
	n := now.In(loc)
	today := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, loc)
	if period == "month" {
		end := time.Date(n.Year(), n.Month(), 1, 0, 0, 0, 0, loc)
		return end.AddDate(0, -1, 0), end
	}
	end := today.AddDate(0, 0, 1)
	return end.AddDate(0, 0, -7), end
}

// reportDue reports whether a report of period should be generated now.
func reportDue(period string, now time.Time, loc *time.Location) bool {
	n := now.In(loc)
	if period == "month" {
		return n.Day() == 1 && n.Hour() >= 10
	}
	return n.Weekday() == time.Friday && n.Hour() >= 18
}

// RunReports generates and sends due weekly/monthly owner reports. Safe to call often.
func (s *Service) RunReports(ctx context.Context) (int, error) {
	clinics, err := s.store.ListOperatingClinics(ctx)
	if err != nil {
		return 0, err
	}
	sent := 0
	var errs []error
	for _, c := range clinics {
		for _, period := range []string{"week", "month"} {
			if period == "week" && !c.Settings.WeeklyReport || period == "month" && !c.Settings.MonthlyReport {
				continue
			}
			if !reportDue(period, s.now(), c.Location()) {
				continue
			}
			ok, err := s.SendReport(ctx, c.ID, period, false)
			if err != nil {
				errs = append(errs, fmt.Errorf("clinic %s %s report: %w", c.ID, period, err))
				continue
			}
			if ok {
				sent++
			}
		}
	}
	return sent, errors.Join(errs...)
}

// SendReport builds the current report for period and sends it to the owner.
// Unless force is set it is sent once per period (idempotent across instances).
func (s *Service) SendReport(ctx context.Context, clinicID uuid.UUID, period string, force bool) (bool, error) {
	sent := false
	err := s.store.InTx(ctx, func(tx *store.Store) error {
		c, err := tx.GetClinic(ctx, clinicID)
		if err != nil {
			return err
		}
		start, end := ReportPeriod(period, s.now(), c.Location())
		if !force {
			exists, err := tx.ReportExists(ctx, c.ID, period, start)
			if err != nil || exists {
				return err
			}
		}
		m, err := tx.ClinicMetrics(ctx, c.ID, start, end, c.AvgCheckKZT)
		if err != nil {
			return err
		}
		msg, err := s.toOwner(ctx, tx, c, domain.MsgReport, RenderReport(c, period, m))
		if err != nil {
			return err
		}
		var msgID *uuid.UUID
		if msg != nil {
			msgID = &msg.ID
		}
		err = tx.InsertReport(ctx, c.ID, period, start, end, m, msgID)
		if errors.Is(err, domain.ErrConflict) {
			if !force {
				return err // another instance won the race: roll back our message
			}
			err = nil
		}
		sent = err == nil
		return err
	})
	if errors.Is(err, domain.ErrConflict) {
		return false, nil
	}
	return sent, err
}

// RenderReport renders the owner report (the mascot «Кезекші» speaks on the clinic side).
func RenderReport(c *domain.Clinic, period string, m *store.Metrics) string {
	loc := c.Location()
	from := m.From.In(loc)
	to := m.To.In(loc).Add(-time.Second)
	title := "за неделю"
	if period == "month" {
		title = "за месяц"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "📊 Кезекші %s (%s – %s)\n«%s»\n\n", title, messaging.FormatDate(from, "ru"), messaging.FormatDate(to, "ru"), c.Name)
	fmt.Fprintf(&b, "• Записей: %d\n", m.Appointments)
	fmt.Fprintf(&b, "• Напоминаний отправлено: %d\n", m.RemindersSent)
	fmt.Fprintf(&b, "• Подтвердили визит: %d\n", m.Confirmed)
	if m.PreventedNoShows > 0 {
		fmt.Fprintf(&b, "• Предупредили заранее вместо неявки: %d (≈ %s ₸ не потеряно)\n", m.PreventedNoShows, money(m.SavedRevenueKZT))
	} else {
		b.WriteString("• Предупредили заранее вместо неявки: 0\n")
	}
	fmt.Fprintf(&b, "• Неявок: %d\n", m.NoShows)
	if m.Leads > 0 {
		fmt.Fprintf(&b, "• Новых заявок: %d%s\n", m.Leads, sourcesSuffix(m.LeadsBySource))
	}
	if m.Reviews > 0 {
		fmt.Fprintf(&b, "• Оценок: %d, средняя %.1f; приглашены в 2ГИС: %d, жалоб лично вам: %d\n", m.Reviews, m.AvgRating, m.ReviewsTo2GIS, m.Complaints)
	}
	if m.BotChats > 0 {
		fmt.Fprintf(&b, "• Пациентов написали в WhatsApp: %d\n", m.BotChats)
	}
	b.WriteString("\n💡 " + reportTip(c, m))
	return b.String()
}

func reportTip(c *domain.Clinic, m *store.Metrics) string {
	switch {
	case m.RemindersSent == 0:
		return "Напоминания не уходили: пришлите Кезекші список записей на завтра или проверьте подключение WhatsApp."
	case m.NoShowRate > 0.1:
		return "Неявок больше 10%. Попросите администратора звонить тем, кто не ответил на напоминание до 10:00."
	case m.Leads == 0 && !domain.PlanAllowsSite(c.Plan):
		return "Вечерние пациенты не могут записаться, пока администратор спит. Сайт с онлайн-записью принимает заявки 24/7 — тариф «Полное кресло»."
	case m.Reviews > 0 && m.ReviewsTo2GIS < 3:
		return "Мало приглашений в 2ГИС: напоминайте пациентам после визита, что им придёт короткий вопрос об оценке."
	case m.RescheduleRequested > 0:
		return "Освободившиеся после переносов окна можно предложить пациентам из листа ожидания."
	default:
		return "Пациенты подтверждают визиты вовремя — кресло не простаивает. Так держать!"
	}
}

func sourcesSuffix(m map[string]int) string {
	if len(m) == 0 {
		return ""
	}
	names := map[string]string{"site": "сайт", "whatsapp": "WhatsApp", "2gis": "2ГИС", "instagram": "Instagram", "other": "другое"}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		n := names[k]
		if n == "" {
			n = k
		}
		parts = append(parts, fmt.Sprintf("%s %d", n, m[k]))
	}
	return " (" + strings.Join(parts, ", ") + ")"
}

// money formats 180000 as "180 000".
func money(n int) string {
	s := fmt.Sprintf("%d", n)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteRune(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}
