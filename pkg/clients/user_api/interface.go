package user_api

import (
	"context"
	"github.com/google/uuid"
)

type Client interface {
	CreateUser(ctx context.Context, req CreateUserRequest) (*UserResponse, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*UserResponse, error)
	FindByEmail(ctx context.Context, email string) (*UserResponse, error)
	AuthenticateUser(ctx context.Context, email, password string) (*UserResponse, error)
}
