// Package notebook parses the free-text schedule an administrator sends to the bot
// ("тетрадь" mode), e.g.
//
//	Завтра:
//	10:00 Алия чистка +7 701 123 45 67
//	11.30 - Серик, 87021234567, пломба
//
// The parser is deterministic (no AI) so the administrator always gets predictable results;
// every line it cannot understand is reported back to them.
package notebook

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/almasik/jazyl/internal/phone"
)

// Entry is one parsed appointment line.
type Entry struct {
	Line     int       `json:"line"`
	StartsAt time.Time `json:"starts_at"`
	Name     string    `json:"name"`
	Phone    string    `json:"phone"`
	Service  string    `json:"service"`
}

// Failure is a line that could not be parsed.
type Failure struct {
	Line   int    `json:"line"`
	Text   string `json:"text"`
	Reason string `json:"reason"`
}

// Result of parsing a whole message.
type Result struct {
	Date     time.Time `json:"date"` // local midnight of the target day
	Entries  []Entry   `json:"entries"`
	Failures []Failure `json:"failures"`
}

var (
	reTime  = regexp.MustCompile(`(?:^|[^\d])([01]?\d|2[0-3])[:.]([0-5]\d)(?:[^\d]|$)`)
	rePhone = regexp.MustCompile(`\+?\d[\d\s\-()]{8,16}\d`)
	reDate  = regexp.MustCompile(`(?:^|[^\d])(0?[1-9]|[12]\d|3[01])\.(0?[1-9]|1[0-2])(?:\.(\d{2,4}))?(?:[^\d]|$)`)
)

var (
	tomorrowWords = []string{"завтра", "ертең", "ертен", "tomorrow"}
	todayWords    = []string{"сегодня", "бүгін", "бугин", "today"}
)

// LooksLikeSchedule is a cheap check used to tell a schedule apart from a casual message.
func LooksLikeSchedule(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		if reTime.MatchString(line) && rePhone.MatchString(line) {
			return true
		}
	}
	return false
}

// Parse parses text relative to now in loc. Default target day is tomorrow.
func Parse(text string, now time.Time, loc *time.Location) Result {
	now = now.In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	res := Result{Date: today.AddDate(0, 0, 1)}

	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for i, raw := range lines {
		lineNo := i + 1
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		phoneMatch := rePhone.FindString(line)

		// Header lines without a phone may set the target date.
		if phoneMatch == "" {
			if d, ok := parseDateHeader(line, today); ok {
				res.Date = d
				continue
			}
			if reTime.MatchString(line) {
				res.Failures = append(res.Failures, Failure{Line: lineNo, Text: line, Reason: "нет номера телефона"})
			}
			// Free text (greetings etc.) is ignored silently.
			continue
		}

		p, err := phone.Normalize(phoneMatch)
		if err != nil {
			res.Failures = append(res.Failures, Failure{Line: lineNo, Text: line, Reason: "неверный номер телефона"})
			continue
		}
		rest := strings.Replace(line, phoneMatch, " ", 1)

		tm := reTime.FindStringSubmatchIndex(rest)
		if tm == nil {
			res.Failures = append(res.Failures, Failure{Line: lineNo, Text: line, Reason: "нет времени (пример 10:30)"})
			continue
		}
		hh, _ := strconv.Atoi(rest[tm[2]:tm[3]]) // regex guarantees digits
		mm, _ := strconv.Atoi(rest[tm[4]:tm[5]])
		rest = rest[:tm[2]] + " " + rest[tm[5]:]

		name, service := splitNameService(rest)
		res.Entries = append(res.Entries, Entry{
			Line:    lineNo,
			Name:    name,
			Phone:   p,
			Service: service,
			// StartsAt is resolved after all headers are read (a date header may come later).
			StartsAt: time.Date(0, 1, 1, hh, mm, 0, 0, loc),
		})
	}

	for i := range res.Entries {
		t := res.Entries[i].StartsAt
		res.Entries[i].StartsAt = time.Date(res.Date.Year(), res.Date.Month(), res.Date.Day(), t.Hour(), t.Minute(), 0, 0, loc)
	}
	return res
}

func parseDateHeader(line string, today time.Time) (time.Time, bool) {
	low := strings.ToLower(line)
	for _, w := range todayWords {
		if strings.Contains(low, w) {
			return today, true
		}
	}
	for _, w := range tomorrowWords {
		if strings.Contains(low, w) {
			return today.AddDate(0, 0, 1), true
		}
	}
	m := reDate.FindStringSubmatch(line)
	if m == nil {
		return time.Time{}, false
	}
	day, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	year := today.Year()
	if m[3] != "" {
		y, _ := strconv.Atoi(m[3])
		if y < 100 {
			y += 2000
		}
		year = y
	}
	d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, today.Location())
	if d.Day() != day { // e.g. 31.02 rolled over
		return time.Time{}, false
	}
	// "05.01" sent in late December means next year.
	if m[3] == "" && d.Before(today.AddDate(0, -1, 0)) {
		d = d.AddDate(1, 0, 0)
	}
	return d, true
}

// splitNameService treats leading capitalised words as the name and the rest as the service.
func splitNameService(s string) (name, service string) {
	s = strings.Map(func(r rune) rune {
		if r == ',' || r == ';' || r == '|' || r == '—' || r == '–' || r == '-' || r == '/' {
			return ' '
		}
		return r
	}, s)
	words := strings.Fields(s)
	var nameParts []string
	i := 0
	for ; i < len(words) && len(nameParts) < 2; i++ {
		r := []rune(words[i])
		if len(r) == 0 || !unicode.IsUpper(r[0]) {
			break
		}
		nameParts = append(nameParts, words[i])
	}
	if len(nameParts) == 0 && len(words) > 0 {
		nameParts = []string{words[0]}
		i = 1
	}
	return strings.Join(nameParts, " "), strings.Join(words[i:], " ")
}
