package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP     HTTPConfig
	GRPC     GRPCConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
	Storage  StorageConfig
}

type HTTPConfig struct {
	Port            string        `env:"PORT"                  env-default:"8083"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT"     env-default:"10s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT"    env-default:"0s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"30s"`
}

type GRPCConfig struct {
	Port            string        `env:"GRPC_PORT"             env-default:"9083"`
	ShutdownTimeout time.Duration `env:"GRPC_SHUTDOWN_TIMEOUT" env-default:"30s"`
}

type DatabaseConfig struct {
	URL          string `env:"PROFILE_DATABASE_URL" env-required:"true"`
	MaxOpenConns int    `env:"DB_MAX_OPEN_CONNS"    env-default:"25"`
	MaxIdleConns int    `env:"DB_MAX_IDLE_CONNS"    env-default:"5"`
}

type KafkaConfig struct {
	BrokersRaw string `env:"KAFKA_BROKERS"  env-default:"kafka:9092"`
	Topic      string `env:"KAFKA_TOPIC"    env-required:"true"`
	GroupID    string `env:"KAFKA_GROUP_ID" env-required:"true"`
}

type StorageConfig struct {
	Endpoint        string `env:"STORAGE_ENDPOINT"         env-required:"true"`
	AccessKeyID     string `env:"STORAGE_ACCESS_KEY_ID"    env-required:"true"`
	SecretAccessKey string `env:"STORAGE_SECRET_ACCESS_KEY" env-required:"true"`
	BucketName      string `env:"STORAGE_BUCKET_NAME"      env-default:"avatars"`
	UseSSL          bool   `env:"STORAGE_USE_SSL"          env-default:"false"`
	PublicBaseURL   string `env:"STORAGE_PUBLIC_BASE_URL"  env-required:"true"`
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
			"max idle connections (%d) cannot exceed max open connections (%d)",
			c.Database.MaxIdleConns, c.Database.MaxOpenConns,
		)
	}
	if len(c.Kafka.Brokers()) == 0 {
		return fmt.Errorf("kafka brokers list is empty")
	}
	return nil
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
