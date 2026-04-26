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
	Redis    RedisConfig
	Kafka    KafkaConfig
	Telegram TelegramConfig
}

type HTTPConfig struct {
	Port            string        `env:"PORT" env-default:"8086"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"30s"`
}

type DatabaseConfig struct {
	URL          string `env:"NOTIFICATION_DATABASE_URL" env-required:"true"`
	MaxOpenConns int    `env:"DB_MAX_OPEN_CONNS" env-default:"25"`
	MaxIdleConns int    `env:"DB_MAX_IDLE_CONNS" env-default:"5"`
}

type RedisConfig struct {
	Addr     string        `env:"REDIS_ADDR" env-default:"localhost:6379"`
	Password string        `env:"REDIS_PASSWORD" env-default:""`
	DB       int           `env:"REDIS_DB" env-default:"2"`
	Timeout  time.Duration `env:"REDIS_TIMEOUT" env-default:"3s"`
}

type KafkaConfig struct {
	BrokersRaw string `env:"KAFKA_BROKERS" env-default:"kafka:9092"`
	GroupID    string `env:"KAFKA_GROUP_ID" env-default:"notification_service"`
}

type TelegramConfig struct {
	BotToken string `env:"TELEGRAM_BOT_TOKEN" env-default:""`
}

func (k KafkaConfig) Brokers() []string {
	result := make([]string, 0)

	for _, b := range strings.Split(k.BrokersRaw, ",") {
		b = strings.TrimSpace(b)
		if b != "" {
			result = append(result, b)
		}
	}

	return result
}

func (t TelegramConfig) Enabled() bool { return t.BotToken != "" }

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
		return fmt.Errorf("max_idle_conns (%d) > max_open_conns (%d)", c.Database.MaxIdleConns, c.Database.MaxOpenConns)
	}

	if len(c.Kafka.Brokers()) == 0 {
		return fmt.Errorf("kafka brokers list is empty")
	}
	return nil
}
