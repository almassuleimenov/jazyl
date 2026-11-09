// Command jazyl runs the Jazyl backend: REST API, OpenWA webhook receiver, reminder scheduler and
// the WhatsApp outbox worker, all in one binary (workers can be disabled per instance).
//
// Usage:
//
//	jazyl                     # serve (API + workers)
//	jazyl migrate             # apply database migrations and exit
//	jazyl create-superadmin EMAIL PASSWORD
//	jazyl reset-mfa EMAIL      # remove a lost 2FA device (server-side recovery)
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
	case "reset-mfa":
		// Recovery for a Jazyl account whose authenticator phone was lost. Run on the server.
		if len(os.Args) != 3 {
			return errors.New("usage: jazyl reset-mfa EMAIL")
		}
		pool, err := platform.OpenDB(ctx, cfg.DatabaseURL, 2)
		if err != nil {
			return err
		}
		defer pool.Close()
		return resetMFA(ctx, store.New(pool), os.Args[2], log)
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

func resetMFA(ctx context.Context, st *store.Store, email string, log *slog.Logger) error {
	u, err := st.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if err := st.DisableTOTP(ctx, u.ID); err != nil {
		return err
	}
	if err := st.RevokeUserTokens(ctx, u.ID); err != nil {
		return err
	}
	if err := st.Audit(ctx, u.ClinicID, nil, "user.totp_reset_cli", "user", u.ID.String(), "", nil); err != nil {
		log.Warn("audit write failed", "err", err)
	}
	log.Info("2FA removed; the user must set it up again at next login", "user_id", u.ID)
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

	handler, err := api.New(api.Deps{
		Config: cfg, Log: log, Store: st, Service: svc, Tokens: tokens, Hub: hub, Sessions: openwa, Router: router,
	})
	if err != nil {
		return err
	}
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
	// The env pair is the founder's login and the source of truth: the account is created if
	// missing, and its password follows the env on every start. 2FA and sessions of a matching
	// password are left untouched; other users are never affected.
	u, err := st.GetUserByEmail(ctx, cfg.BootstrapAdminEmail)
	if errors.Is(err, domain.ErrNotFound) {
		return createSuperadmin(ctx, st, cfg.BootstrapAdminEmail, cfg.BootstrapAdminPassword, log)
	}
	if err != nil {
		return fmt.Errorf("bootstrap admin: %w", err)
	}
	if u.Role != domain.RoleSuperadmin {
		log.Warn("BOOTSTRAP_ADMIN_EMAIL belongs to a clinic user; left unchanged", "user_id", u.ID)
		return nil
	}
	if !u.IsActive {
		log.Warn("bootstrap superadmin is deactivated; left unchanged", "user_id", u.ID)
		return nil
	}
	if auth.CheckPassword(u.PasswordHash, cfg.BootstrapAdminPassword) {
		return nil
	}
	hash, err := auth.HashPassword(cfg.BootstrapAdminPassword)
	if err != nil {
		return fmt.Errorf("bootstrap admin password: %w", err)
	}
	if err := st.SetUserPassword(ctx, u.ID, hash, false); err != nil {
		return err
	}
	// The old password is gone: sign out every session it opened.
	if err := st.RevokeUserTokens(ctx, u.ID); err != nil {
		return err
	}
	if err := st.Audit(ctx, nil, nil, "user.password_from_env", "user", u.ID.String(), "", nil); err != nil {
		log.Warn("audit write failed", "err", err)
	}
	log.Info("superadmin password updated from BOOTSTRAP_ADMIN_PASSWORD", "user_id", u.ID)
	return nil
}
