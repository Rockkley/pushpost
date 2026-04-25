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
	"github.com/redis/go-redis/v9"
	"github.com/rockkley/pushpost/clients/user_api"
	"github.com/rockkley/pushpost/services/auth_service/internal/config"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain/usecase"
	smtpemail "github.com/rockkley/pushpost/services/auth_service/internal/email/smtp"
	redisrepo "github.com/rockkley/pushpost/services/auth_service/internal/repository/redis"
	"github.com/rockkley/pushpost/services/auth_service/internal/transport"
	myHTTP "github.com/rockkley/pushpost/services/auth_service/internal/transport/http"
	"github.com/rockkley/pushpost/services/auth_service/internal/transport/http/middleware"
	"github.com/rockkley/pushpost/services/common_service/jwt"
	"github.com/rockkley/pushpost/services/common_service/logger"
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

	userClient, err := user_api.NewUserClient(cfg.UserSvc.BaseURL, &http.Client{
		Timeout: cfg.UserSvc.Timeout,
	})

	if err != nil {
		appLog.Error("failed to create user client", slog.Any("error", err))
		os.Exit(1)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err = rdb.Ping(context.Background()).Err(); err != nil {
		appLog.Error("failed to connect to Redis", slog.Any("error", err))
		os.Exit(1)
	}
	defer rdb.Close()

	sessionStore := redisrepo.NewSessionStore(rdb, cfg.Redis.Timeout)
	otpStore := redisrepo.NewOTPStore(rdb, cfg.Redis.Timeout)

	emailSender := smtpemail.NewSender(smtpemail.Config{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.User,
		Password: cfg.SMTP.Pass,
		From:     cfg.SMTP.From,
		AppName:  cfg.SMTP.AppName,
	})

	jwtManager := jwt.NewManager(cfg.JWT.Secret, &cfg.JWT.AccessTTL)
	authUsecase := usecase.NewAuthUsecase(userClient, sessionStore, otpStore, emailSender, jwtManager)
	authHandler := myHTTP.NewAuthHandler(authUsecase)
	authMiddleware := middleware.NewAuthMiddleware(authUsecase)
	mux := transport.NewRouter(appLog, authMiddleware, authHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      mux,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		appLog.Info("auth service started", slog.String("port", cfg.HTTP.Port))
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
		appLog.Info("auth service shutting down...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		appLog.Error("graceful shutdown failed", slog.Any("error", err))
		os.Exit(1)
	}

	appLog.Info("auth service stopped")
}
