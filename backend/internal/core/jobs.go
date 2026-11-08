package core

import (
	"context"
	"time"

	"github.com/almasik/jazyl/internal/observability"
)

// RunJobs runs periodic background work until ctx is cancelled:
// reminders every 20s, auto-complete every 5m, owner reports every 10m, housekeeping hourly.
// All jobs are safe to run on several instances at once (SKIP LOCKED / unique constraints).
func (s *Service) RunJobs(ctx context.Context) error {
	reminders := time.NewTicker(20 * time.Second)
	complete := time.NewTicker(5 * time.Minute)
	reports := time.NewTicker(10 * time.Minute)
	house := time.NewTicker(time.Hour)
	defer reminders.Stop()
	defer complete.Stop()
	defer reports.Stop()
	defer house.Stop()

	s.runReminders(ctx) // catch up immediately after start
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-reminders.C:
			s.runReminders(ctx)
		case <-complete.C:
			s.job(ctx, "autocomplete", func(ctx context.Context) error {
				n, err := s.AutoComplete(ctx)
				if n > 0 {
					s.log.Info("auto-completed visits", "count", n)
				}
				return err
			})
		case <-reports.C:
			s.job(ctx, "reports", func(ctx context.Context) error {
				n, err := s.RunReports(ctx)
				if n > 0 {
					s.log.Info("owner reports sent", "count", n)
				}
				return err
			})
		case <-house.C:
			s.job(ctx, "housekeeping", s.store.Housekeeping)
		}
	}
}

func (s *Service) runReminders(ctx context.Context) {
	s.job(ctx, "reminders", func(ctx context.Context) error {
		for range 20 { // drain up to 20 batches per tick
			n, err := s.ProcessDueReminders(ctx, 50)
			if err != nil || n < 50 {
				return err
			}
		}
		return nil
	})
}

func (s *Service) job(ctx context.Context, name string, fn func(context.Context) error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := fn(ctx); err != nil {
		if ctx.Err() == nil {
			s.log.Error("job failed", "job", name, "err", err)
		}
		observability.JobRuns.WithLabelValues(name, "error").Inc()
		return
	}
	observability.JobRuns.WithLabelValues(name, "ok").Inc()
}
