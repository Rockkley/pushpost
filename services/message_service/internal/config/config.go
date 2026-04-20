package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
}

type HTTPConfig struct {
	Port            string        `env:"PORT"                  env-default:"8084"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT"     env-default:"5s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT"    env-default:"10s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"30s"`
}

type DatabaseConfig struct {
	URL          string `env:"MESSAGE_DATABASE_URL" env-required:"true"`
	MaxOpenConns int    `env:"DB_MAX_OPEN_CONNS"    env-default:"25"`
	MaxIdleConns int    `env:"DB_MAX_IDLE_CONNS"    env-default:"5"`
}

type KafkaConfig struct {
	BrokersRaw string `env:"KAFKA_BROKERS" env-required:"true" env-default:"kafka:9092"`
}

func (k KafkaConfig) Brokers() []string {
	brokers := strings.Split(k.BrokersRaw, ",")
	result := make([]string, 0, len(brokers))
	for _, b := range brokers {
		if b = strings.TrimSpace(b); b != "" {
			result = append(result, b)
		}
	}
	return result
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
	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf(
			"max_idle_conns (%d) cannot exceed max_open_conns (%d)",
			c.Database.MaxIdleConns, c.Database.MaxOpenConns,
		)
	}
	if len(c.Kafka.Brokers()) == 0 {
		return fmt.Errorf("kafka brokers list is empty")
	}
	return nil
}
