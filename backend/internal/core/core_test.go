package core

import (
	"testing"
	"time"

	"github.com/almasik/jazyl/internal/domain"
)

var almaty = time.FixedZone("UTC+5", 5*3600)

func at(day, hour, min int) time.Time { return time.Date(2026, 11, day, hour, min, 0, 0, almaty) }

func TestPlanReminders(t *testing.T) {
	s := domain.DefaultSettings()
	tests := []struct {
		name          string
		starts, now   time.Time
		wantDayBefore *time.Time
		wantTwoHours  *time.Time
	}{
		{"normal booking", at(10, 15, 0), at(8, 12, 0), ptr(at(9, 20, 0)), ptr(at(10, 13, 0))},
		{"morning visit: 2h reminder in quiet hours", at(10, 9, 0), at(8, 12, 0), ptr(at(9, 20, 0)), nil},
		{"booked late evening for tomorrow noon: send next morning", at(10, 12, 0), at(9, 22, 30), ptr(at(10, 8, 0)), nil},
		{"booked same day: send asap, 2h skipped (too close)", at(9, 16, 0), at(9, 11, 0), ptr(at(9, 11, 1)), nil},
		{"booked same day far ahead: asap and 2h", at(9, 18, 0), at(9, 10, 0), ptr(at(9, 10, 1)), ptr(at(9, 16, 0))},
		{"too close to visit: nothing", at(9, 13, 0), at(9, 11, 0), nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PlanReminders(tt.starts, tt.now, s, almaty)
			if !eqTime(p.DayBefore, tt.wantDayBefore) {
				t.Errorf("day_before = %v want %v", fmtp(p.DayBefore), fmtp(tt.wantDayBefore))
			}
			if !eqTime(p.TwoHours, tt.wantTwoHours) {
				t.Errorf("two_hours = %v want %v", fmtp(p.TwoHours), fmtp(tt.wantTwoHours))
			}
		})
	}
}

func TestReviewTime(t *testing.T) {
	s := domain.DefaultSettings()
	if got := ReviewTime(at(9, 15, 0), s, almaty); !got.Equal(at(9, 17, 0)) {
		t.Errorf("afternoon: %v", got)
	}
	if got := ReviewTime(at(9, 20, 30), s, almaty); !got.Equal(at(10, 8, 0)) {
		t.Errorf("evening visit should move to next morning: %v", got)
	}
}

func TestClassifyReply(t *testing.T) {
	tests := map[string]Intent{
		"1": IntentConfirm, " 1 ": IntentConfirm, "Да, приду!": IntentConfirm, "иә": IntentConfirm, "👍": IntentConfirm,
		"2": IntentReschedule, "Перенести пожалуйста": IntentReschedule, "я не смогу": IntentReschedule, "жоқ": IntentReschedule,
		"СТОП": IntentOptOut, "старт": IntentOptIn,
		"Сколько стоит чистка?": IntentUnknown, "": IntentUnknown,
	}
	for in, want := range tests {
		if got := ClassifyReply(in); got != want {
			t.Errorf("ClassifyReply(%q) = %s want %s", in, got, want)
		}
	}
}

func TestParseRating(t *testing.T) {
	tests := []struct {
		in   string
		want int
		ok   bool
	}{
		{"5", 5, true}, {"4 спасибо", 4, true}, {"⭐⭐⭐", 3, true}, {"10", 0, false}, {"отлично", 0, false}, {"7", 0, false},
	}
	for _, tt := range tests {
		got, ok := ParseRating(tt.in)
		if got != tt.want || ok != tt.ok {
			t.Errorf("ParseRating(%q) = %d,%v", tt.in, got, ok)
		}
	}
}

func TestComputeSlots(t *testing.T) {
	open, closeT := at(10, 9, 0), at(10, 12, 0)
	busy := [][2]time.Time{{at(10, 9, 30), at(10, 10, 30)}}
	step, dur := 30*time.Minute, 30*time.Minute

	got := ComputeSlots(open, closeT, at(9, 0, 0), step, dur, 1, busy, almaty)
	want := []string{"09:00", "10:30", "11:00", "11:30"}
	if len(got) != len(want) {
		t.Fatalf("slots = %+v", got)
	}
	for i := range want {
		if got[i].Time != want[i] {
			t.Errorf("slot %d = %s want %s", i, got[i].Time, want[i])
		}
	}
	// Two chairs: the busy interval no longer blocks.
	if got := ComputeSlots(open, closeT, at(9, 0, 0), step, dur, 2, busy, almaty); len(got) != 6 {
		t.Errorf("2 chairs: %d slots", len(got))
	}
	// notBefore cuts the morning.
	if got := ComputeSlots(open, closeT, at(10, 11, 0), step, dur, 1, busy, almaty); len(got) != 2 {
		t.Errorf("notBefore: %d slots", len(got))
	}
}

func TestReportPeriod(t *testing.T) {
	s, e := ReportPeriod("week", at(13, 18, 30), almaty)
	if !s.Equal(at(7, 0, 0)) || !e.Equal(at(14, 0, 0)) {
		t.Errorf("week = %v..%v", s, e)
	}
	s, e = ReportPeriod("month", time.Date(2026, 12, 1, 10, 0, 0, 0, almaty), almaty)
	if !s.Equal(time.Date(2026, 11, 1, 0, 0, 0, 0, almaty)) || !e.Equal(time.Date(2026, 12, 1, 0, 0, 0, 0, almaty)) {
		t.Errorf("month = %v..%v", s, e)
	}
}

func TestMoney(t *testing.T) {
	for in, want := range map[int]string{0: "0", 999: "999", 15000: "15 000", 180000: "180 000", 1234567: "1 234 567"} {
		if got := money(in); got != want {
			t.Errorf("money(%d) = %q want %q", in, got, want)
		}
	}
}

func ptr(t time.Time) *time.Time { return &t }

func eqTime(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(*b)
}

func fmtp(t *time.Time) string {
	if t == nil {
		return "<nil>"
	}
	return t.In(almaty).Format("01-02 15:04")
}
