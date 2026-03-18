package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/rockkley/pushpost/services/common_service/logger"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	"github.com/rockkley/pushpost/services/common_service/outbox/kafka"
	postgres2 "github.com/rockkley/pushpost/services/common_service/outbox/postgres"
	stdlog "log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rockkley/pushpost/services/common_service/database"
	"github.com/rockkley/pushpost/services/user_service/internal/config"
	"github.com/rockkley/pushpost/services/user_service/internal/domain/usecase"
	"github.com/rockkley/pushpost/services/user_service/internal/repository/postgres"
	"github.com/rockkley/pushpost/services/user_service/internal/transport"
	myHTTP "github.com/rockkley/pushpost/services/user_service/internal/transport/http"
)

func main() {
	envFile := os.Getenv("ENV_FILE")

	if envFile == "" {
		envFile = ".env"
	}

	if err := godotenv.Load(envFile); err != nil {
		stdlog.Printf("no env file %q found, using runtime environment variables", envFile)
	}

	cfg, err := config.Load()

	if err != nil {
		stdlog.Fatal("failed to load config:", err)
	}

	appLog := logger.SetupLogger(os.Getenv("APP_ENV"))
	slog.SetDefault(appLog)

	db, err := database.Connect(database.Config{
		URL:          cfg.Database.URL,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
	})

	if err != nil {
		appLog.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}

	defer db.Close()

	uow := postgres.NewUnitOfWork(db)
	userUseCase := usecase.NewUserUseCase(uow)
	userHandler := myHTTP.NewUserHandler(userUseCase)
	mux := transport.NewRouter(appLog, userHandler)

	kafkaPublisher := kafka.NewPublisher(cfg.Kafka.Brokers(), appLog)

	defer func() {
		if err = kafkaPublisher.Close(); err != nil {
			appLog.Error("failed to close kafka publisher", slog.Any("error", err))
		}
	}()

	//publisher := &noopPublisher{log: appLog}
	workerRepo := postgres2.NewOutboxRepository(db)

	outboxWorker := outbox.NewWorker(
		workerRepo,
		kafkaPublisher,
		outbox.DefaultWorkerConfig(),
		appLog,
	)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      mux,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go outboxWorker.Run(ctx)

	serverErr := make(chan error, 1)
	go func() {
		appLog.Info("user service started", slog.String("port", cfg.HTTP.Port))
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		appLog.Error("server failed to start", slog.Any("error", err))
		os.Exit(1)
	case <-quit:
		appLog.Info("user service shutting down...")
	}

	// Останавливаем воркер
	cancel()

	// Отдельный ctx только для graceful shutdown HTTP
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLog.Error("graceful shutdown failed", slog.Any("error", err))
		os.Exit(1)
	}

	appLog.Info("user service stopped")
}

type noopPublisher struct {
	log *slog.Logger
}

func (p *noopPublisher) Publish(ctx context.Context, event *outbox.OutboxEvent) error {
	p.log.Info("noop publisher: event skipped",
		slog.String("event_type", event.EventType),
		slog.String("aggregate_id", event.AggregateID),
	)
	return nil
}
