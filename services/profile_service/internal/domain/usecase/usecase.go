package usecase

import (
	"context"
	"errors"
	"github.com/rockkley/pushpost/clients/user_api"
	"github.com/rockkley/pushpost/services/profile_service/internal/domain"
)

type userReader interface {
	GetUserByUsername(ctx context.Context, username string) (*user_api.UserResponse, error)
}

type ProfileUseCase struct {
	userClient userReader
}

func NewProfileUseCase(userClient userReader) *ProfileUseCase {
	return &ProfileUseCase{userClient: userClient}
}

func (u *ProfileUseCase) GetByUsername(ctx context.Context, username string) (*domain.Profile, error) {
	user, err := u.userClient.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, user_api.ErrNotFound) {
			return nil, domain.ErrProfileNotFound
		}
		return nil, err
	}

	return &domain.Profile{
		UserID:    user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}, nil
}
