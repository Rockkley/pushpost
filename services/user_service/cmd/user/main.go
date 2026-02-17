package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/rockkley/pushpost/services/common/config"
	"github.com/rockkley/pushpost/services/common/database"
	"github.com/rockkley/pushpost/services/user_service/internal/domain/usecase"
	"github.com/rockkley/pushpost/services/user_service/internal/repository/postgres"
	"github.com/rockkley/pushpost/services/user_service/internal/transport"
	myHTTP "github.com/rockkley/pushpost/services/user_service/internal/transport/http"
	"log"
	"net/http"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using default variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to open database:", err)
	}

	defer db.Close()

	userRepo := postgres.NewUserRepository(db)
	userUseCase := usecase.NewUserUseCase(userRepo)
	userHandler := myHTTP.NewUserHandler(userUseCase)
	mux := transport.NewRouter(userHandler)

	// HTTP server
	log.Println("UserService is running on", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), mux); err != nil {
		log.Fatal(err)
	}

}
