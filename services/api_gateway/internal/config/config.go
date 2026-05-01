package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP     HTTPConfig
	JWT      JWTConfig
	CORS     CorsConfig
	Services ServicesConfig
}

type HTTPConfig struct {
	Port            string        `env:"PORT"                  env-default:"8000"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT"     env-default:"10s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT"    env-default:"0s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"30s"`
}

type JWTConfig struct {
	Secret string `env:"JWT_SECRET" env-required:"true"`
}

type CorsConfig struct {
	AllowedOriginsRaw string `env:"CORS_ALLOWED_ORIGINS" env-required:"true" env-separator:","`
	MaxAge            int    `env:"CORS_MAX_AGE" env-default:"300"`
}

func (c *CorsConfig) AllowedOrigins() []string {
	if len(c.AllowedOriginsRaw) == 0 {
		return []string{}
	}

	origins := strings.Split(c.AllowedOriginsRaw, ",")
	result := make([]string, 0, len(origins))

	for _, o := range origins {
		if o = strings.TrimSpace(o); o != "" {
			result = append(result, o)
		}
	}

	return result
}

type ServicesConfig struct {
	AuthService         string        `env:"AUTH_SERVICE_URL"          env-required:"true"`
	UserService         string        `env:"USER_SERVICE_URL"          env-required:"true"`
	FriendshipService   string        `env:"FRIENDSHIP_SERVICE_URL"    env-required:"true"`
	ProfileServiceGRPC  string        `env:"PROFILE_SERVICE_GRPC_ADDR" env-required:"true"`
	ProfileService      string        `env:"PROFILE_SERVICE_URL"       env-required:"true"`
	MessageService      string        `env:"MESSAGE_SERVICE_URL"     env-required:"true"`
	PostService         string        `env:"POST_SERVICE_URL"          env-required:"true"`
	NotificationService string        `env:"NOTIFICATION_SERVICE_URL"  env-required:"true"`
	Timeout             time.Duration `env:"UPSTREAM_TIMEOUT"          env-default:"10s"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {

		return nil, fmt.Errorf("config: %w", err)
	}

	if err := cfg.validate(); err != nil {

		return nil, fmt.Errorf("config: validation: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("jwt_secret must be at least 32 characters, got %d", len(c.JWT.Secret))
	}

	if len(c.CORS.AllowedOrigins()) == 0 {
		return fmt.Errorf("cors_allowed_origins must not be empty")
	}

	return nil
}
