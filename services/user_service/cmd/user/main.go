package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/rockkley/pushpost/services/common/database"
	"github.com/rockkley/pushpost/services/user_service/internal/config"
	"github.com/rockkley/pushpost/services/user_service/internal/domain/usecase"
	"github.com/rockkley/pushpost/services/user_service/internal/repository/postgres"
	"github.com/rockkley/pushpost/services/user_service/internal/transport"
	myHTTP "github.com/rockkley/pushpost/services/user_service/internal/transport/http"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := godotenv.Load("services/user_service/.env"); err != nil {
		log.Println("no .env file found, using default variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	db, err := database.Connect(database.Config{
		URL:          cfg.Database.URL,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
	})

	if err != nil {
		log.Fatal("failed to open database:", err)
	}

	defer db.Close()

	userRepo := postgres.NewUserRepository(db)
	userUseCase := usecase.NewUserUseCase(userRepo)
	userHandler := myHTTP.NewUserHandler(userUseCase)
	mux := transport.NewRouter(userHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      mux,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	go func() {
		slog.Info("UserService is running on", cfg.HTTP.Port)
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {

			log.Fatalf("server error: %v", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("UserService shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}

	slog.Info("UserService stopped")
}
