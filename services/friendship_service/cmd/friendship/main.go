package main

import (
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	"github.com/rockkley/pushpost/services/common_service/database"
	"github.com/rockkley/pushpost/services/common_service/logger"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	"github.com/rockkley/pushpost/services/common_service/outbox/kafka"
	outboxpg "github.com/rockkley/pushpost/services/common_service/outbox/postgres"
	//friendshipv1 "github.com/rockkley/pushpost/services/friendship_service/gen/friendship/v1"
	"github.com/rockkley/pushpost/services/friendship_service/internal/config"
	usecase "github.com/rockkley/pushpost/services/friendship_service/internal/domain/usecase"
	repopg "github.com/rockkley/pushpost/services/friendship_service/internal/repository/postgres"
	"github.com/rockkley/pushpost/services/friendship_service/internal/transport"
	//friendgrpc "github.com/rockkley/pushpost/services/friendship_service/internal/transport/grpc"
	friendhttp "github.com/rockkley/pushpost/services/friendship_service/internal/transport/http"
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

	// Database
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

	// Business logic
	uow := repopg.NewUnitOfWork(db)
	friendUseCase := usecase.NewFriendshipUseCase(uow)

	// Kafka outbox worker
	kafkaPublisher := kafka.NewPublisher(cfg.Kafka.Brokers(), appLog)
	defer func() {
		if closeErr := kafkaPublisher.Close(); closeErr != nil {
			appLog.Error("failed to close kafka publisher", slog.Any("error", closeErr))
		}
	}()

	outboxWorker := outbox.NewWorker(
		outboxpg.NewOutboxRepository(db),
		kafkaPublisher,
		outbox.DefaultWorkerConfig(),
		appLog,
	)

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go outboxWorker.Run(workerCtx)

	// HTTP server
	httpHandler := friendhttp.NewFriendshipHandler(friendUseCase)
	mux := router.NewRouter(appLog, httpHandler)

	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      mux,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	// gRPC server
	grpcSrv := grpc.NewServer()
	//friendshipv1.RegisterFriendshipServiceServer(
	//	grpcSrv,
	//	friendgrpc.NewFriendshipServer(friendUseCase, appLog),
	//)

	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPC.Port))
	if err != nil {
		appLog.Error("failed to listen gRPC port", slog.Any("error", err))
		os.Exit(1)
	}

	// Start
	serverErr := make(chan error, 2)

	go func() {
		appLog.Info("friendship HTTP server started", slog.String("port", cfg.HTTP.Port))
		if srvErr := httpSrv.ListenAndServe(); !errors.Is(srvErr, http.ErrServerClosed) {
			serverErr <- srvErr
		}
	}()

	go func() {
		appLog.Info("friendship gRPC server started", slog.String("port", cfg.GRPC.Port))
		if srvErr := grpcSrv.Serve(grpcLis); srvErr != nil {
			serverErr <- srvErr
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err = <-serverErr:
		appLog.Error("server error", slog.Any("error", err))
		os.Exit(1)
	case <-quit:
		appLog.Info("friendship service shutting down...")
	}

	workerCancel()

	grpcSrv.GracefulStop()
	appLog.Info("gRPC server stopped")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer shutdownCancel()
	if err = httpSrv.Shutdown(shutdownCtx); err != nil {
		appLog.Error("HTTP graceful shutdown failed", slog.Any("error", err))
		os.Exit(1)
	}

	appLog.Info("friendship service stopped")
}
