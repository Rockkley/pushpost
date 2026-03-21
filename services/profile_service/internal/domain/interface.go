package domain

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"time"
)

var ErrProfileNotFound = errors.New("profile not found")

type ProfileUseCase interface {
	GetByUsername(ctx context.Context, username string) (*Profile, error)
}

type Profile struct {
	UserID    uuid.UUID
	Username  string
	CreatedAt time.Time
}
