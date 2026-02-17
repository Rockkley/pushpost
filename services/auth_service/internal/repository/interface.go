package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain"
)

type SessionStore interface {
	Save(ctx context.Context, session *domain.Session) error
	Get(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error)
	Delete(ctx context.Context, sessionID uuid.UUID) error
}
