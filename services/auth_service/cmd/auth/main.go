package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/rockkley/pushpost/pkg/clients/user_api"
	"github.com/rockkley/pushpost/pkg/jwt"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain/usecase"
	"github.com/rockkley/pushpost/services/auth_service/internal/repository/memory"
	"github.com/rockkley/pushpost/services/auth_service/internal/transport"
	myHTTP "github.com/rockkley/pushpost/services/auth_service/internal/transport/http"
	"github.com/rockkley/pushpost/services/auth_service/internal/transport/http/middleware"
	"github.com/rockkley/pushpost/services/common/config"
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

	sessionStore := memory.NewSessionStore()
	jwtManager := jwt.NewManager(cfg.JWTSecret)
	userClient, err := user_api.NewUserClient(fmt.Sprintf("http://localhost:%s", cfg.Port), nil)
	if err != nil {
		log.Fatal("failed to create user client:", err)
	}
	authUsecase := usecase.NewAuthUsecase(userClient, sessionStore, jwtManager)
	authHandler := myHTTP.NewAuthHandler(authUsecase)

	authMiddleware := middleware.NewAuthMiddleware(authUsecase)
	mux := transport.NewRouter(authMiddleware, authHandler)

	// HTTP server
	log.Println("AuthService is running on", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", "8081"), mux); err != nil {
		log.Fatal(err)
	}

}
