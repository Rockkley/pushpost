package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/rockkley/pushpost/pkg/clients/user_api"
	"github.com/rockkley/pushpost/pkg/jwt"
	"github.com/rockkley/pushpost/services/auth_service/internal/config"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain/usecase"
	"github.com/rockkley/pushpost/services/auth_service/internal/repository/memory"
	"github.com/rockkley/pushpost/services/auth_service/internal/transport"
	myHTTP "github.com/rockkley/pushpost/services/auth_service/internal/transport/http"
	"github.com/rockkley/pushpost/services/auth_service/internal/transport/http/middleware"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := godotenv.Load("services/auth_service/.env"); err != nil {
		log.Println("no .env file found, using default variables")
	}

	cfg, err := config.Load()

	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	logger := newLogger("development")
	slog.SetDefault(logger)

	userClient, err := user_api.NewUserClient(cfg.UserSvc.BaseURL, &http.Client{
		Timeout: cfg.UserSvc.Timeout,
	})

	if err != nil {
		logger.Error("failed to create user client", slog.Any("error", err))
		os.Exit(1)
	}

	sessionStore := memory.NewSessionStore()
	jwtManager := jwt.NewManager(cfg.JWT.Secret, nil)
	authUsecase := usecase.NewAuthUsecase(userClient, sessionStore, jwtManager)
	authHandler := myHTTP.NewAuthHandler(authUsecase)
	authMiddleware := middleware.NewAuthMiddleware(authUsecase)
	mux := transport.NewRouter(logger, authMiddleware, authHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      mux,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("auth service started", slog.String("port", cfg.HTTP.Port))
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		logger.Error("server failed to start", slog.Any("error", err))
		os.Exit(1)
	case <-quit:
		logger.Info("auth service shutting down...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("auth service stopped")
}

func newLogger(env string) *slog.Logger {
	switch env {
	case "production":
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "development":
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	default: // local
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
}
