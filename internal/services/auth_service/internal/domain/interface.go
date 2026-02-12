package domain

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/services/auth_service/transport/http/dto"
	"github.com/rockkley/pushpost/internal/services/user_service/internal/entity"
)

type AuthUsecase interface {
	Register(ctx context.Context, dto dto.RegisterUserDto) (*entity.User, error)
	Login(ctx context.Context, dto dto.LoginUserDTO) (string, error)
	Logout(ctx context.Context, tokenID uuid.UUID) error
	//GetSessionByToken(ctx context.Context, tokenStr string) (*domain.Session, error)
	AuthenticateRequest(ctx context.Context, tokenStr string) (*Session, error)
}
