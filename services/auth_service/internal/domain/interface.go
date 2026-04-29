package domain

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain/dto"
)

type AuthUsecase interface {
	Register(ctx context.Context, data dto.RegisterUserDTO) (*dto.RegisterResponseDTO, error)
	Login(ctx context.Context, dto dto.LoginUserDTO) (string, error)
	Logout(ctx context.Context, tokenID uuid.UUID) error
	//GetSessionByToken(ctx context.Context, tokenStr string) (*domain.Session, error)
	AuthenticateRequest(ctx context.Context, tokenStr string) (*Session, error)
	VerifyEmail(ctx context.Context, email, code string) error
	ResendOTP(ctx context.Context, email string) error
}
