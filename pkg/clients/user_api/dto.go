package user_api

import "github.com/google/uuid"

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

type AuthenticateUserDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
