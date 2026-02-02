package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/domain"
	"github.com/rockkley/pushpost/internal/handler/http/dto"
)

type AuthService interface {
	Register(ctx context.Context, dto dto.RegisterUserDto) (*domain.User, error)
	Login(ctx context.Context, dto dto.LoginUserDTO) (string, error)
	Logout(ctx context.Context, tokenID uuid.UUID) error
	//GetSessionByToken(ctx context.Context, tokenStr string) (*domain.Session, error)
	AuthenticateRequest(ctx context.Context, tokenStr string) (*domain.Session, error)
}

type MessageService interface {
	SendMessage(ctx context.Context, senderID, receiverID uuid.UUID, content string) (*domain.Message, error)
}
