// Package config loads runtime configuration from environment variables (and an optional .env file).
package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the full runtime configuration of the Jazyl backend.
type Config struct {
	Env      string // "development" | "production"
	HTTPAddr string
	// PublicBaseURL is how OpenWA (and websites) reach this backend, e.g. https://api.jazyl.kz.
	PublicBaseURL string
	LogLevel      string

	DatabaseURL        string
	DatabaseMigrateURL string // direct (non-pooled) URL, needed for session-level advisory locks
	DatabaseMaxConns   int32
	AutoMigrate        bool

	JWTSecret []byte
	// DataEncryptionKey encrypts TOTP secrets at rest (AES-256-GCM, key derived with SHA-256).
	DataEncryptionKey []byte
	// RequireSuperadminMFA keeps superadmin sessions out of everything except account
	// security until they passed TOTP. Defaults to true in production.
	RequireSuperadminMFA bool
	AccessTokenTTL       time.Duration
	RefreshTokenTTL      time.Duration

	BootstrapAdminEmail    string
	BootstrapAdminPassword string

	// CORS for the Jazyl dashboard (frontend). Clinic websites are allowed per site key.
	DashboardOrigins []string

	OpenWABaseURL       string
	OpenWAAPIKey        string
	OpenWAWebhookSecret string

	// Pacing between two outbound messages from the same WhatsApp number (anti-ban).
	SendMinInterval time.Duration
	SendJitter      time.Duration

	WorkersEnabled bool
	TrustedProxy   bool
	// LoginRatePerMinute limits login and 2FA attempts per client IP. Behind the dashboard BFF
	// set TRUSTED_PROXY=true so the real client IP (X-Forwarded-For) is used.
	LoginRatePerMinute int
	MetricsToken       string
}

// Load reads configuration. It loads envFile (if it exists) without overriding real env vars.
func Load(envFiles ...string) (*Config, error) {
	for _, f := range envFiles {
		if err := loadDotEnv(f); err != nil {
			return nil, err
		}
	}

	c := &Config{
		Env:                    get("APP_ENV", "development"),
		HTTPAddr:               get("HTTP_ADDR", ":8080"),
		PublicBaseURL:          strings.TrimRight(get("PUBLIC_BASE_URL", "http://localhost:8080"), "/"),
		LogLevel:               get("LOG_LEVEL", "info"),
		DatabaseURL:            first(os.Getenv("DATABASE_URL"), os.Getenv("POSTRE_DATABASE_API"), os.Getenv("POSTGRES_DATABASE_API")),
		DatabaseMigrateURL:     os.Getenv("DATABASE_MIGRATE_URL"),
		AutoMigrate:            getBool("AUTO_MIGRATE", true),
		AccessTokenTTL:         getDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:        getDuration("REFRESH_TOKEN_TTL", 30*24*time.Hour),
		BootstrapAdminEmail:    os.Getenv("BOOTSTRAP_ADMIN_EMAIL"),
		BootstrapAdminPassword: os.Getenv("BOOTSTRAP_ADMIN_PASSWORD"),
		DashboardOrigins:       splitList(get("DASHBOARD_ORIGINS", "http://localhost:3000")),
		OpenWABaseURL:          strings.TrimRight(get("OPENWA_BASE_URL", "http://localhost:2785"), "/"),
		OpenWAAPIKey:           os.Getenv("OPENWA_API_KEY"),
		OpenWAWebhookSecret:    os.Getenv("OPENWA_WEBHOOK_SECRET"),
		SendMinInterval:        getDuration("SEND_MIN_INTERVAL", 12*time.Second),
		SendJitter:             getDuration("SEND_JITTER", 10*time.Second),
		WorkersEnabled:         getBool("WORKERS_ENABLED", true),
		TrustedProxy:           getBool("TRUSTED_PROXY", false),
		MetricsToken:           os.Getenv("METRICS_TOKEN"),
	}
	c.DatabaseMaxConns = int32(getInt("DATABASE_MAX_CONNS", 10))
	c.LoginRatePerMinute = getInt("LOGIN_RATE_PER_MINUTE", 20)
	if c.DatabaseMigrateURL == "" {
		// Neon: the direct endpoint is the pooled host without the "-pooler" suffix.
		c.DatabaseMigrateURL = strings.Replace(c.DatabaseURL, "-pooler.", ".", 1)
	}
	c.JWTSecret = []byte(os.Getenv("JWT_SECRET"))
	c.DataEncryptionKey = []byte(os.Getenv("DATA_ENCRYPTION_KEY"))
	c.RequireSuperadminMFA = getBool("REQUIRE_SUPERADMIN_MFA", c.IsProduction())

	return c, c.validate()
}

// IsProduction reports whether the app runs in production mode.
func (c *Config) IsProduction() bool { return c.Env == "production" }

func (c *Config) validate() error {
	var errs []error
	if c.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL (or POSTRE_DATABASE_API) is required"))
	}
	if len(c.JWTSecret) < 32 {
		if c.IsProduction() {
			errs = append(errs, errors.New("JWT_SECRET must be at least 32 bytes in production"))
		} else {
			c.JWTSecret = []byte("dev-only-insecure-jwt-secret-change-me-please!!")
		}
	}
	if len(c.DataEncryptionKey) < 32 {
		if c.IsProduction() {
			errs = append(errs, errors.New("DATA_ENCRYPTION_KEY must be at least 32 bytes in production"))
		} else {
			c.DataEncryptionKey = []byte("dev-only-insecure-data-key-change-me-please!!!!!")
		}
	}
	if c.IsProduction() {
		if !c.RequireSuperadminMFA {
			errs = append(errs, errors.New("REQUIRE_SUPERADMIN_MFA cannot be disabled in production"))
		}
		if c.OpenWAWebhookSecret == "" || len(c.OpenWAWebhookSecret) < 16 {
			errs = append(errs, errors.New("OPENWA_WEBHOOK_SECRET (16+ chars) is required in production"))
		}
		if c.OpenWAAPIKey == "" {
			errs = append(errs, errors.New("OPENWA_API_KEY is required in production"))
		}
	}
	return errors.Join(errs...)
}

func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		if _, exists := os.LookupEnv(k); !exists {
			if err := os.Setenv(k, v); err != nil {
				return fmt.Errorf("setenv %s: %w", k, err)
			}
		}
	}
	return sc.Err()
}

func get(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func first(vals ...string) string {
	for _, v := range vals {
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	return ""
}

func getBool(k string, def bool) bool {
	b, err := strconv.ParseBool(os.Getenv(k))
	if err != nil {
		return def
	}
	return b
}

func getInt(k string, def int) int {
	n, err := strconv.Atoi(os.Getenv(k))
	if err != nil {
		return def
	}
	return n
}

func getDuration(k string, def time.Duration) time.Duration {
	d, err := time.ParseDuration(os.Getenv(k))
	if err != nil {
		return def
	}
	return d
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
