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
	"github.com/rockkley/pushpost/clients/user_api"
	"github.com/rockkley/pushpost/services/common_service/logger"
	"github.com/rockkley/pushpost/services/profile_service/internal/config"
	"github.com/rockkley/pushpost/services/profile_service/internal/domain/usecase"
	"github.com/rockkley/pushpost/services/profile_service/internal/transport"
	transportHTTP "github.com/rockkley/pushpost/services/profile_service/internal/transport/http"
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

	userClient, err := user_api.NewUserClient(cfg.UserSvc.BaseURL, &http.Client{Timeout: cfg.UserSvc.Timeout})
	if err != nil {
		appLog.Error("failed to create user client", slog.Any("error", err))
		os.Exit(1)
	}

	uc := usecase.NewProfileUseCase(userClient)
	handler := transportHTTP.NewProfileHandler(uc)
	mux := transport.NewRouter(appLog, handler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      mux,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		appLog.Info("profile service started", slog.String("port", cfg.HTTP.Port))
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err = <-serverErr:
		appLog.Error("profile service failed", slog.Any("error", err))
		os.Exit(1)
	case <-quit:
		appLog.Info("profile service shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err = srv.Shutdown(shutdownCtx); err != nil {
		appLog.Error("graceful shutdown failed", slog.Any("error", err))
		os.Exit(1)
	}
}
