package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/domain"
	"github.com/rockkley/pushpost/internal/repository"
	passwordTools "github.com/rockkley/pushpost/pkg/password"
	"github.com/rockkley/pushpost/pkg/validator"
	"time"
)

type AuthService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(ctx context.Context, username, email, password string) (*domain.User, error) {

	validationResults := validator.ValidateRegisterInputs(username, email, password)
	if len(validationResults.Errors) > 0 {
		return nil, validationResults.Errors[0]
	}

	hashedPassword, err := passwordTools.Hash(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Id:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
