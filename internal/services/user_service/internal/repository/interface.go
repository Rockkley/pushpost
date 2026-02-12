package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/services/user_service/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
