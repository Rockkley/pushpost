package domain

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	"github.com/rockkley/pushpost/services/user_service/internal/domain/dto"
	"github.com/rockkley/pushpost/services/user_service/internal/entity"
	"github.com/rockkley/pushpost/services/user_service/internal/repository"
)

type UserUseCaseInterface interface {
	CreateUser(ctx context.Context, dto dto.CreateUserDTO) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

type Tx interface {
	Users() repository.UserRepositoryInterface
	Outbox() outbox.WriterInterface
}
type UnitOfWorkInterface interface {
	Do(ctx context.Context, fn func(tx Tx) error) error
	Reader() repository.UserRepositoryInterface
}
