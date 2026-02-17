package domain

import (
	"context"
	"github.com/google/uuid"
	dto2 "github.com/rockkley/pushpost/services/auth_service/internal/transport/http/dto"
)

type AuthUsecase interface {
	Register(ctx context.Context, data dto2.RegisterUserDto) (*dto2.RegisterResponseDto, error)
	Login(ctx context.Context, dto dto2.LoginUserDTO) (string, error)
	Logout(ctx context.Context, tokenID uuid.UUID) error
	//GetSessionByToken(ctx context.Context, tokenStr string) (*domain.Session, error)
	AuthenticateRequest(ctx context.Context, tokenStr string) (*Session, error)
}
