package repository

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/user_service/internal/entity"
	"github.com/rockkley/pushpost/services/user_service/internal/outbox"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type OutboxRepository interface {
	Insert(ctx context.Context, tx *sql.Tx, event *outbox.OutboxEvent) error
	FetchPending(ctx context.Context, limit int) ([]*outbox.OutboxEvent, error)
	MarkAsProcessed(ctx context.Context, id uuid.UUID) error
}
