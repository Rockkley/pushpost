package domain

import (
	"context"
	"github.com/rockkley/pushpost/services/user_service/internal/domain/dto"
	"github.com/rockkley/pushpost/services/user_service/internal/entity"
)

type UserUseCase interface {
	AuthenticateUser(ctx context.Context, dto dto.AuthenticateUserRequestDTO) (*entity.User, error)
	CreateUser(ctx context.Context, dto dto.CreateUserDTO) (*entity.User, error)
}
