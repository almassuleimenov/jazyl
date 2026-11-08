package messaging

import (
	"strings"
	"testing"
	"time"
)

func TestDayBeforeHasNoDiagnosisAndAsksReply(t *testing.T) {
	v := Visit{Lang: "ru", PatientName: "Алия", ClinicName: "Дент-Люкс", StartsAt: time.Date(2026, 11, 10, 15, 0, 0, 0, time.UTC)}
	s := DayBefore(v)
	for _, want := range []string{"Алия", "«Дент-Люкс»", "10 ноября", "15:00", "1 — приду", "2 — нужно перенести"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in %q", want, s)
		}
	}
	kk := DayBefore(Visit{Lang: "kk", ClinicName: "Дент", StartsAt: v.StartsAt})
	if !strings.Contains(kk, "қараша") || !strings.Contains(kk, "1 — келемін") {
		t.Errorf("kk template: %q", kk)
	}
}

func TestBackoff(t *testing.T) {
	tests := []struct {
		n    int
		want time.Duration
	}{{0, 30 * time.Second}, {1, 30 * time.Second}, {2, time.Minute}, {4, 4 * time.Minute}, {20, 30 * time.Minute}}
	for _, tt := range tests {
		if got := Backoff(tt.n); got != tt.want {
			t.Errorf("Backoff(%d) = %s want %s", tt.n, got, tt.want)
		}
	}
}

func TestScheduleAck(t *testing.T) {
	d := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	s := ScheduleAck(d, []ScheduleLine{{At: d.Add(10 * time.Hour), Name: "Алия", New: true}}, []string{"строка 3"}, 20)
	if !strings.Contains(s, "Принято записей на 10 ноября: 1") || !strings.Contains(s, "Не понял") {
		t.Errorf("ack = %q", s)
	}
}
