package main

import (
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	goredis "github.com/redis/go-redis/v9"
	"github.com/rockkley/pushpost/services/common_service/database"
	"github.com/rockkley/pushpost/services/common_service/logger"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	kafkap "github.com/rockkley/pushpost/services/common_service/outbox/kafka"
	outboxpg "github.com/rockkley/pushpost/services/common_service/outbox/postgres"
	"github.com/rockkley/pushpost/services/post_service/internal/cache/redis"
	"github.com/rockkley/pushpost/services/post_service/internal/clients/friendship"
	"github.com/rockkley/pushpost/services/post_service/internal/config"
	"github.com/rockkley/pushpost/services/post_service/internal/domain/usecase"
	repopg "github.com/rockkley/pushpost/services/post_service/internal/repository/postgres"
	"github.com/rockkley/pushpost/services/post_service/internal/transport"
	myHTTP "github.com/rockkley/pushpost/services/post_service/internal/transport/http"
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

	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err = rdb.Ping(context.Background()).Err(); err != nil {
		appLog.Error("failed to connect to redis", slog.Any("error", err))
		os.Exit(1)
	}
	defer rdb.Close()

	friendshipClient, err := friendship.NewGRPCClient(cfg.Friendship.GRPCAddr)

	if err != nil {
		appLog.Error("failed to create friendship grpc client", slog.Any("error", err))
		os.Exit(1)
	}

	feedCache := redis.NewFeedCache(rdb)
	uow := repopg.NewUnitOfWork(db)
	uc := usecase.NewPostUseCase(uow, friendshipClient, feedCache)
	handler := myHTTP.NewPostHandler(uc)
	mux := transport.NewRouter(appLog, handler)

	kafkaPublisher := kafkap.NewPublisher(cfg.Kafka.Brokers(), appLog)
	defer kafkaPublisher.Close()

	outboxWorker := outbox.NewWorker(
		outboxpg.NewOutboxRepository(db),
		kafkaPublisher,
		outbox.DefaultWorkerConfig(),
		appLog,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go outboxWorker.Run(ctx)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      mux,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	serverErr := make(chan error, 1)

	go func() {
		appLog.Info("post service started", slog.String("port", cfg.HTTP.Port))
		if err = srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err = <-serverErr:
		appLog.Error("server failed to start", slog.Any("error", err))
		os.Exit(1)
	case <-quit:
		appLog.Info("post service shutting down...")
	}

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer shutdownCancel()

	if err = srv.Shutdown(shutdownCtx); err != nil {
		appLog.Error("graceful shutdown failed", slog.Any("error", err))
		os.Exit(1)
	}

	appLog.Info("post service stopped")
}
