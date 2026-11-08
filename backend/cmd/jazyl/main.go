// Command jazyl runs the Jazyl backend: REST API, OpenWA webhook receiver, reminder scheduler and
// the WhatsApp outbox worker, all in one binary (workers can be disabled per instance).
//
// Usage:
//
//	jazyl                     # serve (API + workers)
//	jazyl migrate             # apply database migrations and exit
//	jazyl create-superadmin EMAIL PASSWORD
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	_ "time/tzdata" // embed tz database: Asia/Almaty must resolve on minimal images

	"github.com/almasik/jazyl/internal/api"
	"github.com/almasik/jazyl/internal/auth"
	"github.com/almasik/jazyl/internal/config"
	"github.com/almasik/jazyl/internal/core"
	"github.com/almasik/jazyl/internal/domain"
	"github.com/almasik/jazyl/internal/messaging"
	"github.com/almasik/jazyl/internal/platform"
	"github.com/almasik/jazyl/internal/realtime"
	"github.com/almasik/jazyl/internal/store"
	"github.com/almasik/jazyl/internal/whatsapp"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(".env", "../.env")
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	log := platform.NewLogger(cfg.LogLevel, !cfg.IsProduction())
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cmd := "serve"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "migrate":
		return migrate(ctx, cfg, log)
	case "create-superadmin":
		if len(os.Args) != 4 {
			return errors.New("usage: jazyl create-superadmin EMAIL PASSWORD")
		}
		pool, err := platform.OpenDB(ctx, cfg.DatabaseURL, 2)
		if err != nil {
			return err
		}
		defer pool.Close()
		return createSuperadmin(ctx, store.New(pool), os.Args[2], os.Args[3], log)
	case "serve":
		return serve(ctx, cfg, log)
	case "healthcheck": // for Docker HEALTHCHECK on distroless images (no curl)
		addr := cfg.HTTPAddr
		if len(addr) > 0 && addr[0] == ':' {
			addr = "127.0.0.1" + addr
		}
		hc := &http.Client{Timeout: 3 * time.Second}
		resp, err := hc.Get("http://" + addr + "/healthz")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("healthz returned %d", resp.StatusCode)
		}
		return nil
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func migrate(ctx context.Context, cfg *config.Config, log *slog.Logger) error {
	pool, err := platform.OpenDB(ctx, cfg.DatabaseMigrateURL, 2)
	if err != nil {
		return fmt.Errorf("open migrate db: %w", err)
	}
	defer pool.Close()
	return platform.Migrate(ctx, pool, log)
}

func createSuperadmin(ctx context.Context, st *store.Store, email, password string, log *slog.Logger) error {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	u, err := st.CreateUser(ctx, &domain.User{Email: email, PasswordHash: hash, Name: "Jazyl", Role: domain.RoleSuperadmin})
	if err != nil {
		return fmt.Errorf("create superadmin: %w", err)
	}
	log.Info("superadmin created", "user_id", u.ID, "email", u.Email)
	return nil
}

func serve(ctx context.Context, cfg *config.Config, log *slog.Logger) error {
	if cfg.AutoMigrate {
		if err := migrate(ctx, cfg, log); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	pool, err := platform.OpenDB(ctx, cfg.DatabaseURL, cfg.DatabaseMaxConns)
	if err != nil {
		return err
	}
	defer pool.Close()
	st := store.New(pool)

	if err := bootstrapAdmin(ctx, cfg, st, log); err != nil {
		return err
	}

	openwa := whatsapp.NewOpenWA(cfg.OpenWABaseURL, cfg.OpenWAAPIKey)
	router := whatsapp.NewRouter(map[string]whatsapp.Sender{
		"openwa": openwa,
		"mock":   &whatsapp.Mock{Log: log},
	})
	hub := realtime.NewHub(log)
	svc := core.New(st, hub, log)
	tokens := auth.NewTokens(cfg.JWTSecret)

	handler := api.New(api.Deps{
		Config: cfg, Log: log, Store: st, Service: svc, Tokens: tokens, Hub: hub, Sessions: openwa, Router: router,
	})
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	var wg sync.WaitGroup
	workerCtx, cancelWorkers := context.WithCancel(ctx)
	defer cancelWorkers()
	if cfg.WorkersEnabled {
		outbox := messaging.NewOutbox(st, router, log, cfg.SendMinInterval, cfg.SendJitter)
		startWorker(workerCtx, &wg, log, "outbox", outbox.Run)
		startWorker(workerCtx, &wg, log, "jobs", svc.RunJobs)
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("http server listening", "addr", cfg.HTTPAddr, "env", cfg.Env, "workers", cfg.WorkersEnabled)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("http server: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("http shutdown", "err", err)
	}
	cancelWorkers()
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-shutdownCtx.Done():
		log.Warn("workers did not stop in time")
	}
	log.Info("bye")
	return nil
}

func startWorker(ctx context.Context, wg *sync.WaitGroup, log *slog.Logger, name string, fn func(context.Context) error) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if v := recover(); v != nil {
				log.Error("worker panicked", "worker", name, "panic", v)
			}
		}()
		log.Info("worker started", "worker", name)
		if err := fn(ctx); err != nil {
			log.Error("worker stopped with error", "worker", name, "err", err)
		}
	}()
}

func bootstrapAdmin(ctx context.Context, cfg *config.Config, st *store.Store, log *slog.Logger) error {
	if cfg.BootstrapAdminEmail == "" || cfg.BootstrapAdminPassword == "" {
		return nil
	}
	n, err := st.CountSuperadmins(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return createSuperadmin(ctx, st, cfg.BootstrapAdminEmail, cfg.BootstrapAdminPassword, log)
}
