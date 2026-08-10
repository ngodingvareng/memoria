package config

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort         string        `mapstructure:"SERVER_PORT"`
	SecureCookies      bool          `mapstructure:"SECURE_COOKIES"`
	JWTSecret          string        `mapstructure:"JWT_SECRET"`
	JWTIssuer          string        `mapstructure:"JWT_ISSUER"`
	JWTAccessTokenTTL  time.Duration `mapstructure:"JWT_ACCESS_TOKEN_TTL"`
	JWTRefreshTokenTTL time.Duration `mapstructure:"JWT_REFRESH_TOKEN_TTL"`
	// DBHost defaults to "localhost" for local dev (`make dev`) and
	// testing, but must be set to the Postgres service/host name (e.g.
	// "postgres") when the app runs in its own container, as it does in
	// docker-compose.yml — the app and database are no longer reachable
	// on the same loopback address at that point.
	DBHost              string `mapstructure:"DATABASE_HOST"`
	DBUsername          string `mapstructure:"DATABASE_USERNAME"`
	DBPassword          string `mapstructure:"DATABASE_PASSWORD"`
	DBName              string `mapstructure:"DATABASE_DBNAME"`
	DBPort              string `mapstructure:"DATABASE_PORT"`
	StorageEndpoint     string `mapstructure:"STORAGE_ENDPOINT"`
	StorageRegion       string `mapstructure:"STORAGE_REGION"`
	StorageAccessKey    string `mapstructure:"STORAGE_ACCESS_KEY"`
	StorageSecretKey    string `mapstructure:"STORAGE_SECRET_KEY"`
	StorageBucket       string `mapstructure:"STORAGE_BUCKET"`
	StorageUsePathStyle bool   `mapstructure:"STORAGE_USE_PATH_STYLE"`

	// SMTP* configures outbound email (currently just the
	// password-reset link) — any SMTP server works, not a specific
	// vendor's API. SMTPUsername/SMTPPassword may be left empty for a
	// local dev catch-all (e.g. Mailpit/MailHog) that accepts
	// unauthenticated mail.
	SMTPHost     string `mapstructure:"SMTP_HOST"`
	SMTPPort     string `mapstructure:"SMTP_PORT"`
	SMTPUsername string `mapstructure:"SMTP_USERNAME"`
	SMTPPassword string `mapstructure:"SMTP_PASSWORD"`
	SMTPFrom     string `mapstructure:"SMTP_FROM"`

	// WebBaseURL is the frontend's own origin, used to build links that
	// go into emails (e.g. the password-reset link) — the API has no
	// other way to know where the web app is served from.
	WebBaseURL string `mapstructure:"WEB_BASE_URL"`

	// CORSAllowedOrigins is comma-separated in .env (e.g.
	// "http://localhost:5173,https://app.example.com") — viper's default
	// mapstructure decode hooks (the same ones that parse
	// JWT_ACCESS_TOKEN_TTL's "5m" into a time.Duration) split it into a
	// []string automatically.
	CORSAllowedOrigins []string `mapstructure:"CORS_ALLOWED_ORIGINS"`

	// LoginRateLimitMax/Window throttle POST /auth/login and
	// /auth/register per client IP, independently of the per-account
	// lockout below — this is what stops a single IP from hammering
	// many different accounts (or brute-forcing one) with request
	// volume alone.
	LoginRateLimitMax    int           `mapstructure:"LOGIN_RATE_LIMIT_MAX"`
	LoginRateLimitWindow time.Duration `mapstructure:"LOGIN_RATE_LIMIT_WINDOW"`

	// LoginMaxFailedAttempts/LockoutDuration implement per-account
	// lockout: after this many consecutive failed password attempts
	// against one credential account, it's locked for LockoutDuration
	// regardless of which IP is trying it.
	LoginMaxFailedAttempts int           `mapstructure:"LOGIN_MAX_FAILED_ATTEMPTS"`
	LoginLockoutDuration   time.Duration `mapstructure:"LOGIN_LOCKOUT_DURATION"`
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUsername, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("SERVER_PORT", "3000")
	viper.SetDefault("DATABASE_HOST", "localhost")
	viper.SetDefault("LOGIN_RATE_LIMIT_MAX", 10)
	viper.SetDefault("LOGIN_RATE_LIMIT_WINDOW", time.Minute)
	viper.SetDefault("LOGIN_MAX_FAILED_ATTEMPTS", 5)
	viper.SetDefault("LOGIN_LOCKOUT_DURATION", 15*time.Minute)
	viper.SetDefault("WEB_BASE_URL", "http://localhost:5173")

	if err := viper.ReadInConfig(); err != nil {
		// A missing .env is fine — env vars alone are a valid config
		// source in production. Anything else (e.g. a malformed file)
		// should stay visible rather than silently falling back to
		// zero-value config fields.
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			slog.Warn("failed to read .env config file", "error", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
