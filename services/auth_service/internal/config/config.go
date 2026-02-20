package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type Config struct {
	HTTP    HTTPConfig
	JWT     JWTConfig
	UserSvc UserServiceConfig
}

type HTTPConfig struct {
	Port            string        `env:"PORT"                  env-default:"8081"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT"     env-default:"5s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT"    env-default:"10s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"30s"`
}

type JWTConfig struct {
	Secret    string        `env:"JWT_SECRET"     env-required:"true"`
	AccessTTL time.Duration `env:"JWT_ACCESS_TTL" env-default:"24h"`
}

type UserServiceConfig struct {
	BaseURL string        `env:"USER_SERVICE_URL"     env-required:"true"`
	Timeout time.Duration `env:"USER_SERVICE_TIMEOUT" env-default:"5s"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf(
			"jwt_secret must be at least 32 characters, got %d",
			len(c.JWT.Secret),
		)
	}
	return nil
}
