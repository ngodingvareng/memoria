package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort    string `mapstructure:"SERVER_PORT"`
	SecureCookies bool   `mapstructure:"SECURE_COOKIES"`

	DBUsername string `mapstructure:"DATABASE_USERNAME"`
	DBPassword string `mapstructure:"DATABASE_PASSWORD"`
	DBName     string `mapstructure:"DATABASE_DBNAME"`
	DBPort     string `mapstructure:"DATABASE_PORT"`

	StorageEndpoint     string `mapstructure:"STORAGE_ENDPOINT"`
	StorageRegion       string `mapstructure:"STORAGE_REGION"`
	StorageAccessKey    string `mapstructure:"STORAGE_ACCESS_KEY"`
	StorageSecretKey    string `mapstructure:"STORAGE_SECRET_KEY"`
	StorageBucket       string `mapstructure:"STORAGE_BUCKET"`
	StorageUsePathStyle bool   `mapstructure:"STORAGE_USE_PATH_STYLE"`
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable",
		c.DBUsername, c.DBPassword, c.DBPort, c.DBName)
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Load defaults if you like
	viper.SetDefault("SERVER_PORT", "3000")

	if err := viper.ReadInConfig(); err != nil {
		// Log if needed, or ignore if using env vars
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
