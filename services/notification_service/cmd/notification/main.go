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
	"github.com/rockkley/pushpost/services/notification_service/internal/config"
	"github.com/rockkley/pushpost/services/notification_service/internal/delivery"
	"github.com/rockkley/pushpost/services/notification_service/internal/delivery/inapp"
	tg "github.com/rockkley/pushpost/services/notification_service/internal/delivery/telegram"
	"github.com/rockkley/pushpost/services/notification_service/internal/domain/usecase"
	notifkafka "github.com/rockkley/pushpost/services/notification_service/internal/kafka"
	repopg "github.com/rockkley/pushpost/services/notification_service/internal/repository/postgres"
	redisrepo "github.com/rockkley/pushpost/services/notification_service/internal/repository/redis"
	"github.com/rockkley/pushpost/services/notification_service/internal/transport"
	myHTTP "github.com/rockkley/pushpost/services/notification_service/internal/transport/http"
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

	notifRepo := repopg.NewNotificationRepository(db)
	prefRepo := repopg.NewPreferenceRepository(db)
	telegramRepo := repopg.NewTelegramRepository(db)
	linkStore := redisrepo.NewLinkCodeStore(rdb)

	binder := usecase.NewTelegramBinder(linkStore, telegramRepo, appLog)

	deliverers := []delivery.Deliverer{inapp.NewDeliverer(rdb)}
	var bot *tg.Bot
	if cfg.Telegram.Enabled() {
		bot, err = tg.NewBot(cfg.Telegram.BotToken, binder, appLog)
		if err != nil {
			appLog.Error("failed to create telegram bot", slog.Any("error", err))
			os.Exit(1)
		}
		deliverers = append(deliverers, tg.NewDeliverer(telegramRepo, bot))
	}

	uc := usecase.NewNotificationUseCase(notifRepo, prefRepo, binder, deliverers, appLog)

	handlers := notifkafka.NewHandlers(uc, appLog)
	router := notifkafka.NewRouter(handlers, appLog)

	// notifkafka.ConsumedTopics — единственный источник правды для списка топиков.
	// Не дублируем список здесь, чтобы не допустить рассинхронизации с router.
	consumer := notifkafka.NewConsumer(
		cfg.Kafka.Brokers(),
		cfg.Kafka.GroupID,
		notifkafka.ConsumedTopics,
		router,
		appLog,
	)

	handler := myHTTP.NewNotificationHandler(uc)
	sseHandler := myHTTP.NewSSEHandler(rdb)
	mux := transport.NewRouter(appLog, handler, sseHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      mux,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverErr := make(chan error, 2)

	go func() {
		if runErr := consumer.Run(ctx); runErr != nil {
			serverErr <- fmt.Errorf("kafka consumer: %w", runErr)
		}
	}()

	if bot != nil {
		go func() {
			appLog.Info("telegram bot started")
			if runErr := bot.Run(ctx); runErr != nil {
				appLog.Error("telegram bot stopped with error", slog.Any("error", runErr))
			}
		}()
	}

	go func() {
		appLog.Info("notification service started", slog.String("port", cfg.HTTP.Port))
		if runErr := srv.ListenAndServe(); !errors.Is(runErr, http.ErrServerClosed) {
			serverErr <- runErr
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err = <-serverErr:
		appLog.Error("service error", slog.Any("error", err))
		os.Exit(1)
	case <-quit:
		appLog.Info("notification service shutting down...")
	}

	cancel()
	_ = consumer.Close()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer shutdownCancel()

	if err = srv.Shutdown(shutdownCtx); err != nil {
		appLog.Error("graceful shutdown failed", slog.Any("error", err))
		os.Exit(1)
	}

	appLog.Info("notification service stopped")
}
