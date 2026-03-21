package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP    HTTPConfig
	UserSvc UserServiceConfig
}

type HTTPConfig struct {
	Port            string        `env:"PORT" env-default:"8083"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"30s"`
}

type UserServiceConfig struct {
	BaseURL string        `env:"USER_SERVICE_URL" env-required:"true"`
	Timeout time.Duration `env:"USER_SERVICE_TIMEOUT" env-default:"5s"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	return &cfg, nil
}
