package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	DBUsername string `mapstructure:"DATABASE_USERNAME"`
	DBPassword string `mapstructure:"DATABASE_PASSWORD"`
	DBName     string `mapstructure:"DATABASE_DBNAME"`
	DBPort     string `mapstructure:"DATABASE_PORT"`

	ServerPort string `mapstructure:"SERVER_PORT"`
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
