package domain

import (
	"context"
	"errors"
	"github.com/rockkley/pushpost/services/profile_service/internal/entity"
)

var ErrProfileNotFound = errors.New("profile not found")

type ProfileUseCaseInterface interface {
	GetByUsername(ctx context.Context, username string) (*entity.Profile, error)
	CreateProfile(ctx context.Context, profile *entity.Profile) error
}
