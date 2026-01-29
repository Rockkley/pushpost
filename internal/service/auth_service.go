package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/domain"
	"github.com/rockkley/pushpost/internal/handler/http/dto"
	"github.com/rockkley/pushpost/internal/repository"
	passwordTools "github.com/rockkley/pushpost/pkg/password"
)

type AuthService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(ctx context.Context, dto dto.RegisterUserDto) (*domain.User, error) {

	hashedPassword, err := passwordTools.Hash(dto.Password)
	if err != nil {

		return nil, err
	}

	user := &domain.User{
		Id:           uuid.New(),
		Username:     dto.Username,
		Email:        dto.Email,
		PasswordHash: hashedPassword,
	}

	err = s.userRepo.Create(ctx, user)

	var de domain.DomainError

	if err != nil {
		if errors.As(err, &de) {

			return nil, err
		}

		return nil, fmt.Errorf("internal error - %w", err)
	}
	return user, nil
}
