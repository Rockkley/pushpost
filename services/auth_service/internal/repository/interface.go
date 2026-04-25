package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain"
	"time"
)

type SessionStore interface {
	Save(ctx context.Context, session *domain.Session) error
	Get(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error)
	Delete(ctx context.Context, sessionID uuid.UUID) error
}

type OTPStore interface {
	Save(ctx context.Context, email, code string, ttl time.Duration) error
	Get(ctx context.Context, email string) (string, error)
	Delete(ctx context.Context, email string) error
	IncrAttempts(ctx context.Context, email string, ttl time.Duration) (int64, error)
	SetCooldown(ctx context.Context, email string, ttl time.Duration) error
	HasCooldown(ctx context.Context, email string) (bool, error)
}
