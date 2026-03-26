package app

import (
	"context"
	"log/slog"

	"github.com/rockkley/pushpost/services/profile_service/internal/config"
	"github.com/rockkley/pushpost/services/profile_service/internal/kafka"
)

type App struct {
	consumer *kafka.Consumer
	log      *slog.Logger
}

func New(cfg config.Config, log *slog.Logger) *App {
	userCreatedHandler := kafka.NewUserCreatedProcessor(log)
	router := kafka.NewRouter(userCreatedHandler, log)

	consumer := kafka.NewConsumer(
		cfg.KafkaBrokers,
		cfg.KafkaTopic,
		cfg.KafkaGroupID,
		router,
		log,
	)

	return &App{
		consumer: consumer,
		log:      log,
	}
}

func (a *App) Run(ctx context.Context) error {
	return a.consumer.Run(ctx)
}

func (a *App) Close() error {
	return a.consumer.Close()
}
