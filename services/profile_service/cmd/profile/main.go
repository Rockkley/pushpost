package main

import (
	"context"
	"github.com/rockkley/pushpost/services/common_service/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rockkley/pushpost/services/profile_service/internal/app"
	"github.com/rockkley/pushpost/services/profile_service/internal/config"
)

func main() {
	appLog := logger.SetupLogger(os.Getenv("APP_ENV"))
	slog.SetDefault(appLog)
	cfg := config.Load()

	a := app.New(cfg, appLog)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err := a.Run(ctx); err != nil {
		appLog.Error("service stopped with error", err)
	}

	_ = a.Close()
}
