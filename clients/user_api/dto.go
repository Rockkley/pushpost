package user_api

import (
	"github.com/google/uuid"
	"time"
)

type CreateUserRequest struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

type UserResponse struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func (u *UserResponse) IsActive() bool { return u.Status == "active" }
