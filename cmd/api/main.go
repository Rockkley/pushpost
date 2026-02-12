package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/rockkley/pushpost/internal/config"
	"github.com/rockkley/pushpost/internal/database"
	myhttp "github.com/rockkley/pushpost/internal/handler/http"
	"github.com/rockkley/pushpost/internal/handler/http/middleware"
	"github.com/rockkley/pushpost/internal/service/services"
	"github.com/rockkley/pushpost/internal/services/auth_service/repository/memory"
	http2 "github.com/rockkley/pushpost/internal/services/auth_service/transport/http"
	"github.com/rockkley/pushpost/pkg/jwt"
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

	sessionStore := memory.NewSessionStore()
	jwtManager := jwt.NewManager(cfg.JWTSecret)
	userRepo := postgres.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, sessionStore, jwtManager)
	authHandler := http2.NewAuthHandler(authService)

	authMiddleware := middleware.NewAuthMiddleware(authService)
	mux := myhttp.NewRouter(authMiddleware, authHandler)

	// HTTP server
	log.Println("HTTP server is running on", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf("localhost:%s", cfg.Port), mux); err != nil {
		log.Fatal(err)
	}
}
