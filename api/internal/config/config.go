package config

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort          string        `mapstructure:"SERVER_PORT"`
	SecureCookies       bool          `mapstructure:"SECURE_COOKIES"`
	JWTSecret           string        `mapstructure:"JWT_SECRET"`
	JWTIssuer           string        `mapstructure:"JWT_ISSUER"`
	JWTAccessTokenTTL   time.Duration `mapstructure:"JWT_ACCESS_TOKEN_TTL"`
	JWTRefreshTokenTTL  time.Duration `mapstructure:"JWT_REFRESH_TOKEN_TTL"`
	DBUsername          string        `mapstructure:"DATABASE_USERNAME"`
	DBPassword          string        `mapstructure:"DATABASE_PASSWORD"`
	DBName              string        `mapstructure:"DATABASE_DBNAME"`
	DBPort              string        `mapstructure:"DATABASE_PORT"`
	StorageEndpoint     string        `mapstructure:"STORAGE_ENDPOINT"`
	StorageRegion       string        `mapstructure:"STORAGE_REGION"`
	StorageAccessKey    string        `mapstructure:"STORAGE_ACCESS_KEY"`
	StorageSecretKey    string        `mapstructure:"STORAGE_SECRET_KEY"`
	StorageBucket       string        `mapstructure:"STORAGE_BUCKET"`
	StorageUsePathStyle bool          `mapstructure:"STORAGE_USE_PATH_STYLE"`
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable",
		c.DBUsername, c.DBPassword, c.DBPort, c.DBName)
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("SERVER_PORT", "3000")

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
