package main

import (
	"github.com/joho/godotenv"
	"github.com/rockkley/pushpost/internal/config"
	"github.com/rockkley/pushpost/internal/database"
	myhttp "github.com/rockkley/pushpost/internal/handler/http"
	"github.com/rockkley/pushpost/internal/repository/memory"
	"github.com/rockkley/pushpost/internal/repository/postgres"
	"github.com/rockkley/pushpost/internal/service/services"
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
	authHandler := myhttp.NewAuthHandler(authService)

	authMiddleware := myhttp.NewAuthMiddleware(authService)
	mux := myhttp.NewRouter(authMiddleware, authHandler)

	// HTTP server
	log.Println("HTTP server is running on", ":8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
