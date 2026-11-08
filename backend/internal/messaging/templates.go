// Package messaging renders WhatsApp texts and runs the outbound outbox worker.
//
// Rules from the product spec: patient messages go out on behalf of the clinic (not Jazyl),
// never mention diagnoses or treatment, and use only "1 / 2"-style replies (no free AI chat).
package messaging

import (
	"fmt"
	"strings"
	"time"
)

var ruMonths = [...]string{"января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря"}
var kkMonths = [...]string{"қаңтар", "ақпан", "наурыз", "сәуір", "мамыр", "маусым", "шілде", "тамыз", "қыркүйек", "қазан", "қараша", "желтоқсан"}

// FormatDate renders "10 ноября" / "10 қараша".
func FormatDate(t time.Time, lang string) string {
	if lang == "kk" {
		return fmt.Sprintf("%d %s", t.Day(), kkMonths[t.Month()-1])
	}
	return fmt.Sprintf("%d %s", t.Day(), ruMonths[t.Month()-1])
}

// FormatTime renders "15:00".
func FormatTime(t time.Time) string { return t.Format("15:04") }

// Visit holds what patient templates need. Times must already be in the clinic location.
type Visit struct {
	Lang        string
	PatientName string
	ClinicName  string
	Address     string
	StartsAt    time.Time
}

func greet(v Visit) string {
	name := strings.TrimSpace(v.PatientName)
	if v.Lang == "kk" {
		if name == "" {
			return "Сәлеметсіз бе!"
		}
		return "Сәлеметсіз бе, " + name + "!"
	}
	if name == "" {
		return "Здравствуйте!"
	}
	return "Здравствуйте, " + name + "!"
}

func addr(v Visit, label string) string {
	if strings.TrimSpace(v.Address) == "" {
		return ""
	}
	return "\n" + label + v.Address
}

// DayBefore is the evening reminder that asks for a 1/2 reply.
func DayBefore(v Visit) string {
	d, t := FormatDate(v.StartsAt, v.Lang), FormatTime(v.StartsAt)
	if v.Lang == "kk" {
		return fmt.Sprintf("%s\n«%s» клиникасы ертең, %s, сағат %s-де сізді күтеді.%s\n\nЖауап беріңіз:\n1 — келемін\n2 — уақытты ауыстыру керек",
			greet(v), v.ClinicName, d, t, addr(v, "Мекенжай: "))
	}
	return fmt.Sprintf("%s\nНапоминаем: вы записаны в «%s» завтра, %s, в %s.%s\n\nОтветьте цифрой:\n1 — приду\n2 — нужно перенести",
		greet(v), v.ClinicName, d, t, addr(v, "Адрес: "))
}

// TwoHours is the same-day reminder. If already confirmed it does not ask for a reply.
func TwoHours(v Visit, confirmed bool) string {
	t := FormatTime(v.StartsAt)
	if v.Lang == "kk" {
		if confirmed {
			return fmt.Sprintf("«%s» сізді бүгін сағат %s-де күтеді. Кездескенше!", v.ClinicName, t)
		}
		return fmt.Sprintf("%s\n«%s» сізді бүгін сағат %s-де күтеді.\n1 — келемін\n2 — уақытты ауыстыру керек", greet(v), v.ClinicName, t)
	}
	if confirmed {
		return fmt.Sprintf("Ждём вас сегодня в %s в «%s». До встречи!", t, v.ClinicName)
	}
	return fmt.Sprintf("%s\nСегодня в %s ждём вас в «%s».\n1 — приду\n2 — нужно перенести", greet(v), t, v.ClinicName)
}

// SameDay is used when a booking is made too late for the evening reminder.
func SameDay(v Visit) string {
	d, t := FormatDate(v.StartsAt, v.Lang), FormatTime(v.StartsAt)
	if v.Lang == "kk" {
		return fmt.Sprintf("%s\nСіз «%s» клиникасына %s, сағат %s-ге жазылдыңыз.\n1 — келемін\n2 — уақытты ауыстыру керек", greet(v), v.ClinicName, d, t)
	}
	return fmt.Sprintf("%s\nВы записаны в «%s» на %s, %s.\n1 — приду\n2 — нужно перенести", greet(v), v.ClinicName, d, t)
}

// ConfirmAck answers "1".
func ConfirmAck(v Visit) string {
	if v.Lang == "kk" {
		return fmt.Sprintf("Рақмет! Сізді %s, сағат %s-де күтеміз.", FormatDate(v.StartsAt, v.Lang), FormatTime(v.StartsAt))
	}
	return fmt.Sprintf("Спасибо! Ждём вас %s в %s.", FormatDate(v.StartsAt, v.Lang), FormatTime(v.StartsAt))
}

// RescheduleAck answers "2".
func RescheduleAck(lang string) string {
	if lang == "kk" {
		return "Жақсы, әкімші сізге хабарласып, басқа уақыт ұсынады."
	}
	return "Хорошо, администратор свяжется с вами и подберёт другое время."
}

// UnknownReply nudges a patient that answered something other than 1/2.
func UnknownReply(lang string) string {
	if lang == "kk" {
		return "Цифрамен жауап беріңіз: 1 — келемін, 2 — уақытты ауыстыру керек."
	}
	return "Пожалуйста, ответьте цифрой: 1 — приду, 2 — нужно перенести."
}

// ReviewRequest asks for a 1–5 rating after a visit.
func ReviewRequest(v Visit) string {
	if v.Lang == "kk" {
		return fmt.Sprintf("«%s» клиникасын таңдағаныңызға рақмет! Келуіңізді 1-ден 5-ке дейін бағалаңызшы (5 — өте жақсы).", v.ClinicName)
	}
	return fmt.Sprintf("Спасибо, что выбрали «%s»! Оцените, пожалуйста, визит от 1 до 5 (5 — отлично).", v.ClinicName)
}

// ReviewGood thanks a happy patient and links the 2GIS card.
func ReviewGood(lang, url string) string {
	if lang == "kk" {
		if url == "" {
			return "Жоғары бағаңызға рақмет!"
		}
		return "Жоғары бағаңызға рақмет! 2ГИС-те пікір қалдырсаңыз, қуанамыз: " + url
	}
	if url == "" {
		return "Спасибо за высокую оценку!"
	}
	return "Спасибо за высокую оценку! Будем благодарны за отзыв в 2ГИС: " + url
}

// ReviewBad asks an unhappy patient for details that go to the owner privately.
func ReviewBad(lang string) string {
	if lang == "kk" {
		return "Өкінішті. Не ұнамағанын жазыңызшы — хабарламаңызды клиника басшысы жеке оқиды."
	}
	return "Нам жаль, что что-то пошло не так. Напишите, пожалуйста, что именно — сообщение получит лично руководитель клиники."
}

// ReviewBadThanks closes the complaint flow.
func ReviewBadThanks(lang string) string {
	if lang == "kk" {
		return "Рақмет, хабарламаңыз басшыға жіберілді. Сізбен хабарласады."
	}
	return "Спасибо, ваше сообщение передано руководителю. С вами свяжутся."
}

// OptOutAck confirms unsubscribing.
func OptOutAck(lang string) string {
	if lang == "kk" {
		return "Сіз еске салғыштардан бас тарттыңыз. Қайта қосу үшін СТАРТ деп жазыңыз."
	}
	return "Вы отписались от напоминаний. Чтобы снова получать их, напишите СТАРТ."
}

// OptInAck confirms re-subscribing.
func OptInAck(lang string) string {
	if lang == "kk" {
		return "Еске салғыштар қайта қосылды."
	}
	return "Напоминания снова включены."
}

// BookingConfirmed is sent when a website booking was auto-confirmed into a slot.
func BookingConfirmed(v Visit) string {
	d, t := FormatDate(v.StartsAt, v.Lang), FormatTime(v.StartsAt)
	if v.Lang == "kk" {
		return fmt.Sprintf("%s\nСіз «%s» клиникасына %s, сағат %s-ге жазылдыңыз. Алдын ала еске саламыз.", greet(v), v.ClinicName, d, t)
	}
	return fmt.Sprintf("%s\nВы записаны в «%s» на %s, %s. Накануне пришлём напоминание.", greet(v), v.ClinicName, d, t)
}

// ---------- staff-facing (Russian; staff chats are internal) ----------

// AdminReschedule tells the administrator that a patient asked to move the visit.
func AdminReschedule(name, phone string, at time.Time) string {
	return fmt.Sprintf("🔁 Пациент %s (+%s) просит перенести запись %s в %s. Позвоните, чтобы подобрать время — окно освободилось.",
		orDash(name), phone, FormatDate(at, "ru"), FormatTime(at))
}

// AdminNewLead tells the administrator about a website booking request.
func AdminNewLead(source, name, phone, service string, preferred *time.Time, comment string) string {
	var b strings.Builder
	label := map[string]string{"site": "с сайта", "whatsapp": "из WhatsApp", "2gis": "из 2ГИС", "instagram": "из Instagram"}[source]
	if label == "" {
		label = "новая"
	}
	fmt.Fprintf(&b, "🆕 Заявка %s: %s, +%s", label, orDash(name), phone)
	if service != "" {
		fmt.Fprintf(&b, "\nУслуга: %s", service)
	}
	if preferred != nil {
		fmt.Fprintf(&b, "\nЖелаемое время: %s в %s", FormatDate(*preferred, "ru"), FormatTime(*preferred))
	}
	if c := strings.TrimSpace(comment); c != "" {
		fmt.Fprintf(&b, "\nКомментарий: %s", c)
	}
	return b.String()
}

// AdminAutoBooked tells the administrator a website booking was placed into the schedule.
func AdminAutoBooked(name, phone string, at time.Time) string {
	return fmt.Sprintf("✅ Запись с сайта: %s (+%s) на %s в %s. Пациенту отправлено подтверждение.",
		orDash(name), phone, FormatDate(at, "ru"), FormatTime(at))
}

// OwnerComplaint forwards a low rating to the owner.
func OwnerComplaint(name, phone string, rating int, visit time.Time, comment string) string {
	s := fmt.Sprintf("⚠️ Низкая оценка после визита %s: %s (+%s) — %d/5.", FormatDate(visit, "ru"), orDash(name), phone, rating)
	if c := strings.TrimSpace(comment); c != "" {
		s += "\nКомментарий пациента: " + c
	}
	return s + "\nОтзыв не ушёл в 2ГИС. Свяжитесь с пациентом лично."
}

// ScheduleLine is one accepted row in the import acknowledgement.
type ScheduleLine struct {
	At   time.Time
	Name string
	New  bool
}

// ScheduleAck summarises a notebook import back to the administrator.
func ScheduleAck(date time.Time, lines []ScheduleLine, failures []string, eveningHour int) string {
	var b strings.Builder
	if len(lines) == 0 {
		b.WriteString("❌ Не нашёл ни одной записи.\n")
	} else {
		fmt.Fprintf(&b, "✅ Принято записей на %s: %d\n", FormatDate(date, "ru"), len(lines))
		for _, l := range lines {
			mark := ""
			if !l.New {
				mark = " (уже была)"
			}
			fmt.Fprintf(&b, "%s %s%s\n", FormatTime(l.At), orDash(l.Name), mark)
		}
		fmt.Fprintf(&b, "Напоминания уйдут в %d:00 накануне и за 2 часа до приёма.\n", eveningHour)
	}
	if len(failures) > 0 {
		b.WriteString("\n⚠️ Не понял строки:\n")
		for _, f := range failures {
			b.WriteString("• " + f + "\n")
		}
		b.WriteString("\nФормат строки: 10:30 Алия +77011234567 чистка")
	}
	return strings.TrimSpace(b.String())
}

// AdminHelp is the reply to "помощь".
func AdminHelp() string {
	return strings.Join([]string{
		"Кезекші на связи 👋 Что я умею:",
		"",
		"📋 Пришлите список записей — каждая строка: время, имя, телефон, услуга:",
		"Завтра",
		"10:00 Алия +77011234567 чистка",
		"11:30 Серик 87021234567",
		"(первая строка — «завтра», «сегодня» или дата 12.11; по умолчанию — завтра)",
		"",
		"📊 «отчёт» — цифры за последние 7 дней",
		"📅 «завтра» / «сегодня» — список записей и статусы",
	}, "\n")
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}
