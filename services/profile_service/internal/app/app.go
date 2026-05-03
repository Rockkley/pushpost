package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/rockkley/pushpost/services/common_service/database"
	profilev1 "github.com/rockkley/pushpost/services/profile_service/gen/profile/v1"
	"github.com/rockkley/pushpost/services/profile_service/internal/config"
	"github.com/rockkley/pushpost/services/profile_service/internal/domain/usecase"
	"github.com/rockkley/pushpost/services/profile_service/internal/kafka"
	repopg "github.com/rockkley/pushpost/services/profile_service/internal/repository/postgres"
	miniostg "github.com/rockkley/pushpost/services/profile_service/internal/storage/minio"
	httptransport "github.com/rockkley/pushpost/services/profile_service/internal/transport"
	grpctransport "github.com/rockkley/pushpost/services/profile_service/internal/transport/grpc"
	profilehttp "github.com/rockkley/pushpost/services/profile_service/internal/transport/http"
)

type App struct {
	consumer  *kafka.Consumer
	grpcSrv   *grpc.Server
	grpcAddr  string
	httpSrv   *http.Server
	httpAddr  string
	httpClose func(context.Context) error
	log       *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger) (*App, error) {
	db, err := database.Connect(database.Config{
		URL:          cfg.Database.URL,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
	})

	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}

	// Object storage
	minioStorage, err := miniostg.New(miniostg.Config{
		Endpoint:        cfg.Storage.Endpoint,
		AccessKeyID:     cfg.Storage.AccessKeyID,
		SecretAccessKey: cfg.Storage.SecretAccessKey,
		BucketName:      cfg.Storage.BucketName,
		UseSSL:          cfg.Storage.UseSSL,
		PublicBaseURL:   cfg.Storage.PublicBaseURL,
	})

	if err != nil {
		return nil, fmt.Errorf("init object storage: %w", err)
	}

	if err = minioStorage.EnsureBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("ensure storage bucket: %w", err)
	}

	profileRepo := repopg.NewProfileRepository(db)
	uc := usecase.NewProfileUseCase(profileRepo, *minioStorage, minioStorage.KeyFromURL)

	userCreatedProcessor := kafka.NewUserCreatedProcessor(uc, log)
	router := kafka.NewRouter(userCreatedProcessor, log)
	consumer := kafka.NewConsumer(
		cfg.Kafka.Brokers(),
		cfg.Kafka.Topic,
		cfg.Kafka.GroupID,
		router,
		log,
	)

	grpcSrv := grpc.NewServer()
	profilev1.RegisterProfileServiceServer(grpcSrv, grpctransport.NewProfileServer(uc, log))

	httpHandler := profilehttp.NewProfileHandler(uc)
	httpRouter := httptransport.NewRouter(log, httpHandler)
	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      httpRouter,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	return &App{
		consumer:  consumer,
		grpcSrv:   grpcSrv,
		grpcAddr:  fmt.Sprintf(":%s", cfg.GRPC.Port),
		httpSrv:   httpSrv,
		httpAddr:  fmt.Sprintf(":%s", cfg.HTTP.Port),
		httpClose: httpSrv.Shutdown,
		log:       log,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", a.grpcAddr)

	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}

	serverErr := make(chan error, 2)

	go func() {
		a.log.Info("profile gRPC server started", slog.String("addr", a.grpcAddr))
		if err := a.grpcSrv.Serve(lis); err != nil {
			serverErr <- err
		}
	}()

	go func() {
		a.log.Info("profile HTTP server started", slog.String("addr", a.httpAddr))

		if err = a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	go func() {
		a.log.Info("profile kafka consumer started")
		if err = a.consumer.Run(ctx); err != nil {
			serverErr <- err
		}
	}()

	select {
	case err = <-serverErr:
		return err
	case <-ctx.Done():
		a.log.Info("profile service shutting down")
		a.grpcSrv.GracefulStop()

		if err = a.httpClose(context.Background()); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}

		return nil
	}
}

func (a *App) Close() error {
	return a.consumer.Close()
}
