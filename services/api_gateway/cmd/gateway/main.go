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
	"github.com/rockkley/pushpost/clients/friendship_api"
	"github.com/rockkley/pushpost/clients/profile_grpc"
	"github.com/rockkley/pushpost/services/api_gateway/internal/config"
	gwmiddleware "github.com/rockkley/pushpost/services/api_gateway/internal/middleware"
	"github.com/rockkley/pushpost/services/api_gateway/internal/proxy"
	"github.com/rockkley/pushpost/services/api_gateway/internal/transport"
	myHTTP "github.com/rockkley/pushpost/services/api_gateway/internal/transport/http"
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

	timeout := cfg.Services.Timeout

	authProxy, err := proxy.New(cfg.Services.AuthService, timeout)
	if err != nil {
		appLog.Error("failed to create auth proxy", slog.Any("error", err))
		os.Exit(1)
	}

	userProxy, err := proxy.NewStrippingAuth(cfg.Services.UserService, timeout)
	if err != nil {
		appLog.Error("failed to create user proxy", slog.Any("error", err))
		os.Exit(1)
	}

	friendshipProxy, err := proxy.NewStrippingAuth(cfg.Services.FriendshipService, timeout)
	if err != nil {
		appLog.Error("failed to create friendship proxy", slog.Any("error", err))
		os.Exit(1)
	}

	messageProxy, err := proxy.NewStrippingAuth(cfg.Services.MessageService, timeout)
	if err != nil {
		appLog.Error("failed to create message proxy", slog.Any("error", err))
		os.Exit(1)
	}

	profileClient, err := profile_grpc.NewClient(cfg.Services.ProfileServiceGRPC)
	if err != nil {
		appLog.Error("failed to create profile grpc client", slog.Any("error", err))
		os.Exit(1)
	}

	friendshipClient, err := friendship_api.NewFriendshipClient(
		cfg.Services.FriendshipService,
		&http.Client{Timeout: timeout},
	)
	if err != nil {
		appLog.Error("failed to create friendship client", slog.Any("error", err))
		os.Exit(1)
	}

	jwtManager := jwt.NewManager(cfg.JWT.Secret, nil)
	authMW := gwmiddleware.NewAuthMiddleware(jwtManager)
	profileHandler := myHTTP.NewProfileHandler(profileClient, friendshipClient)

	mux := transport.NewRouter(appLog, authMW, transport.Proxies{
		Auth:       authProxy,
		User:       userProxy,
		Friendship: friendshipProxy,
		Message:    messageProxy,
	}, *profileHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      mux,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		appLog.Info("api gateway started", slog.String("port", cfg.HTTP.Port))
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err = <-serverErr:
		appLog.Error("gateway failed to start", slog.Any("error", err))
		os.Exit(1)
	case <-quit:
		appLog.Info("api gateway shutting down...")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer shutdownCancel()

	if err = srv.Shutdown(shutdownCtx); err != nil {
		appLog.Error("graceful shutdown failed", slog.Any("error", err))
		os.Exit(1)
	}

	appLog.Info("api gateway stopped")
}
