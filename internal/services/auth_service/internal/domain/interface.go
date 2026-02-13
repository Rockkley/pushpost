package domain

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/services/auth_service/internal/transport/http/dto"
)

type AuthUsecase interface {
	Register(ctx context.Context, data dto.RegisterUserDto) (*dto.RegisterResponseDto, error)
	Login(ctx context.Context, dto dto.LoginUserDTO) (string, error)
	Logout(ctx context.Context, tokenID uuid.UUID) error
	//GetSessionByToken(ctx context.Context, tokenStr string) (*domain.Session, error)
	AuthenticateRequest(ctx context.Context, tokenStr string) (*Session, error)
}
