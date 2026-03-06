package outbox

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type OutboxRepository interface {
	Insert(ctx context.Context, tx *sql.Tx, event *OutboxEvent) error
	ClaimPending(ctx context.Context, limit int, maxAttempts int) ([]*OutboxEvent, error)
	MarkAsProcessed(ctx context.Context, id uuid.UUID) error
	ResetStuck(ctx context.Context, stuckAfter time.Duration) error
	IncrementAttempts(ctx context.Context, id uuid.UUID) error
}

type Publisher interface {
	Publish(ctx context.Context, event *OutboxEvent) error
}
