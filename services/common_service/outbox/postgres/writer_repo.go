package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/rockkley/pushpost/services/common_service/database"
	"github.com/rockkley/pushpost/services/common_service/outbox"
)

type WriterRepository struct {
	exec database.Executor
}

func NewWriterRepository(exec database.Executor) *WriterRepository {
	return &WriterRepository{exec: exec}
}

func (r *WriterRepository) Insert(ctx context.Context, event *outbox.OutboxEvent) error {
	slog.Debug("WriterRepository Insert")
	id := event.ID
	if id == uuid.Nil {
		id = uuid.New()
	}

	_, err := r.exec.ExecContext(ctx, `
		INSERT INTO outbox_events
			(id, aggregate_id, aggregate_type, event_type, payload, status, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`,
		id,
		event.AggregateID,
		event.AggregateType,
		event.EventType,
		event.Payload,
		outbox.StatusPending,
	)

	if err != nil {

		return fmt.Errorf("outbox insert: %w", err)
	}

	return nil
}
