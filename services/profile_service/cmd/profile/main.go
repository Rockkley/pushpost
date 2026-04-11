package main

import (
	"context"
	stdlog "log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rockkley/pushpost/services/common_service/logger"
	"github.com/rockkley/pushpost/services/profile_service/internal/app"
	"github.com/rockkley/pushpost/services/profile_service/internal/config"
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

	a, err := app.New(cfg, appLog)

	if err != nil {
		appLog.Error("failed to initialize app", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err = a.Run(ctx); err != nil {
		appLog.Error("service stopped with error", slog.Any("error", err))
		os.Exit(1)
	}

	if err = a.Close(); err != nil {
		appLog.Error("failed to close app", slog.Any("error", err))
	}

	appLog.Info("profile service stopped")
}
