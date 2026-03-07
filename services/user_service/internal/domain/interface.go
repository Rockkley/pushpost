package domain

import (
	"context"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	"github.com/rockkley/pushpost/services/user_service/internal/domain/dto"
	"github.com/rockkley/pushpost/services/user_service/internal/entity"
	"github.com/rockkley/pushpost/services/user_service/internal/repository"
)

type UserUseCaseInterface interface {
	AuthenticateUser(ctx context.Context, dto dto.AuthenticateUserRequestDTO) (*entity.User, error)
	CreateUser(ctx context.Context, dto dto.CreateUserDTO) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
}

type Tx interface {
	Users() repository.UserRepositoryInterface
	Outbox() outbox.WriterInterface
}
type UnitOfWorkInterface interface {
	Do(ctx context.Context, fn func(tx Tx) error) error
}
