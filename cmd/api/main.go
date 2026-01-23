package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/rockkley/pushpost/internal/config"
	"github.com/rockkley/pushpost/internal/database"
	"github.com/rockkley/pushpost/pkg/password"
	"github.com/rockkley/pushpost/pkg/validator"
	"log"
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

	fmt.Println("database connected OK")

	//tests
	hashed, err := password.Hash("megapassword")

	if err != nil {
		log.Fatal("failed password hashing:", err)
	}

	fmt.Println("password hashed OK", hashed)

	email := "rockkley94@gmail.com"
	if !validator.IsValidEmail(email) {
		log.Fatal("email is not valid - ", email)
	}

	fmt.Println("email validation OK")
	fmt.Println("ALL SYSTEMS WORKS")

}
