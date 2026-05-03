package domain

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/profile_service/internal/entity"
	"io"
)

var ErrProfileNotFound = errors.New("profile not found")

type ProfileUseCaseInterface interface {
	GetByUsername(ctx context.Context, username string) (*entity.Profile, error)
	CreateProfile(ctx context.Context, profile *entity.Profile) error
	UpdateProfile(ctx context.Context, profile *entity.Profile) error
	UploadAvatar(ctx context.Context, userID uuid.UUID, r io.Reader, size int64, contentType string) (string, error)
}
