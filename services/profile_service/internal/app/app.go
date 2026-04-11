package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"github.com/rockkley/pushpost/services/common_service/database"
	profilev1 "github.com/rockkley/pushpost/services/profile_service/gen/profile/v1"
	"github.com/rockkley/pushpost/services/profile_service/internal/config"
	"github.com/rockkley/pushpost/services/profile_service/internal/domain/usecase"
	"github.com/rockkley/pushpost/services/profile_service/internal/kafka"
	repopg "github.com/rockkley/pushpost/services/profile_service/internal/repository/postgres"
	grpctransport "github.com/rockkley/pushpost/services/profile_service/internal/transport/grpc"
)

type App struct {
	consumer *kafka.Consumer
	grpcSrv  *grpc.Server
	grpcAddr string
	log      *slog.Logger
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

	profileRepo := repopg.NewProfileRepository(db)
	uc := usecase.NewProfileUseCase(profileRepo)

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

	return &App{
		consumer: consumer,
		grpcSrv:  grpcSrv,
		grpcAddr: fmt.Sprintf(":%s", cfg.GRPC.Port),
		log:      log,
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
		a.log.Info("profile kafka consumer started")
		if err := a.consumer.Run(ctx); err != nil {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:

		return err
	case <-ctx.Done():
		a.log.Info("profile service shutting down")
		a.grpcSrv.GracefulStop()

		return nil
	}
}

func (a *App) Close() error {
	return a.consumer.Close()
}
