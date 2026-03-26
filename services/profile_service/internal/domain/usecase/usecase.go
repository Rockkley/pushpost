package usecase

import (
	"context"
	"errors"
	"github.com/rockkley/pushpost/clients/user_api"
	"github.com/rockkley/pushpost/services/profile_service/internal/domain"
	"github.com/rockkley/pushpost/services/profile_service/internal/entity"
	"strings"
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

func (u *ProfileUseCase) GetByUsername(ctx context.Context, username string) (*entity.Profile, error) {
	user, err := u.userClient.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, user_api.ErrNotFound) {
			normalized := strings.ToLower(strings.TrimSpace(username))
			if normalized != "" && normalized != username {
				user, retryErr := u.userClient.GetUserByUsername(ctx, normalized)

				if retryErr == nil {

					return &entity.Profile{
						UserID:    user.ID,
						Username:  user.Username,
						CreatedAt: user.CreatedAt,
					}, nil
				}

				if !errors.Is(retryErr, user_api.ErrNotFound) {

					return nil, retryErr
				}
			}

			return nil, domain.ErrProfileNotFound
		}

		return nil, err
	}

	return &entity.Profile{
		UserID:    user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}, nil
}
