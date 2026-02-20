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

	userClient, err := user_api.NewUserClient(cfg.UserSvc.BaseURL, &http.Client{
		Timeout: cfg.UserSvc.Timeout,
	})
	if err != nil {
		log.Fatal("failed to create user client:", err)
	}

	sessionStore := memory.NewSessionStore()
	jwtManager := jwt.NewManager(cfg.JWT.Secret, nil)
	authUsecase := usecase.NewAuthUsecase(userClient, sessionStore, jwtManager)
	authHandler := myHTTP.NewAuthHandler(authUsecase)
	authMiddleware := middleware.NewAuthMiddleware(authUsecase)
	mux := transport.NewRouter(authMiddleware, authHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      mux,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	go func() {
		slog.Info("auth service started", "port", cfg.HTTP.Port)
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error:%v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("auth service shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown error: %v", err)
	}

	slog.Info("auth service stopped")
}
