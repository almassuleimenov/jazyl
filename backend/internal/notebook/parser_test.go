package notebook

import (
	"testing"
	"time"
)

var loc = time.FixedZone("UTC+5", 5*3600)

func TestParse(t *testing.T) {
	now := time.Date(2026, 11, 9, 19, 30, 0, 0, loc)
	text := "Завтра:\n10:00 Алия чистка +7 701 123 45 67\n11.30 - Серик, 87021234567, пломба\n14:00 Айгерим\nпривет\n15:00 Нурлан +7123"

	res := Parse(text, now, loc)
	if got := res.Date.Format("2006-01-02"); got != "2026-11-10" {
		t.Fatalf("date = %s", got)
	}
	if len(res.Entries) != 2 {
		t.Fatalf("entries = %d (%+v)", len(res.Entries), res.Entries)
	}
	e := res.Entries[0]
	if e.Name != "Алия" || e.Phone != "77011234567" || e.Service != "чистка" || e.StartsAt.Format("15:04") != "10:00" {
		t.Errorf("entry0 = %+v", e)
	}
	e = res.Entries[1]
	if e.Name != "Серик" || e.Phone != "77021234567" || e.Service != "пломба" || e.StartsAt.Format("2006-01-02 15:04") != "2026-11-10 11:30" {
		t.Errorf("entry1 = %+v", e)
	}
	// Line 4 has no phone; line 6 has a too-short number (not recognised as a phone).
	if len(res.Failures) != 2 || res.Failures[0].Line != 4 || res.Failures[1].Line != 6 {
		t.Errorf("failures = %+v", res.Failures)
	}
}

func TestParseDateHeader(t *testing.T) {
	now := time.Date(2026, 12, 28, 12, 0, 0, 0, loc)
	tests := []struct{ text, want string }{
		{"сегодня\n16:00 Алия 87011234567", "2026-12-28"},
		{"Ертең\n16:00 Алия 87011234567", "2026-12-29"},
		{"03.01\n16:00 Алия 87011234567", "2027-01-03"},
		{"30.12.2026\n16:00 Алия 87011234567", "2026-12-30"},
		{"16:00 Алия 87011234567", "2026-12-29"},
	}
	for _, tt := range tests {
		res := Parse(tt.text, now, loc)
		if len(res.Entries) != 1 {
			t.Fatalf("%q: entries=%d", tt.text, len(res.Entries))
		}
		if got := res.Entries[0].StartsAt.Format("2006-01-02"); got != tt.want {
			t.Errorf("%q: got %s want %s", tt.text, got, tt.want)
		}
	}
}

func TestLooksLikeSchedule(t *testing.T) {
	if !LooksLikeSchedule("10:00 Алия 87011234567") {
		t.Error("expected schedule")
	}
	if LooksLikeSchedule("Здравствуйте, во сколько вы работаете?") {
		t.Error("expected not schedule")
	}
}

func FuzzParse(f *testing.F) {
	f.Add("Завтра\n10:00 Алия +77011234567 чистка")
	f.Fuzz(func(t *testing.T, s string) {
		_ = Parse(s, time.Now(), loc) // must never panic
	})
}
