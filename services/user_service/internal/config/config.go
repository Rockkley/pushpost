package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
}

type HTTPConfig struct {
	Port            string        `env:"PORT"                  env-default:"8080"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT"     env-default:"5s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT"    env-default:"10s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"30s"`
}

type DatabaseConfig struct {
	URL          string `env:"USER_DATABASE_URL" env-required:"true"`
	MaxOpenConns int    `env:"DB_MAX_OPEN_CONNS" env-default:"25"`
	MaxIdleConns int    `env:"DB_MAX_IDLE_CONNS" env-default:"5"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("config: failed to read environment variables: %w", err)
	}

	if err := cfg.validate(&cfg); err != nil {
		return nil, fmt.Errorf("config: validation error: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate(cfg *Config) error {
	if cfg.Database.MaxIdleConns > cfg.Database.MaxOpenConns {
		return fmt.Errorf(
			"max idle connections (%d) cannot exceed max open connections (%d)",
			c.Database.MaxIdleConns, c.Database.MaxOpenConns,
		)
	}

	return nil
}
