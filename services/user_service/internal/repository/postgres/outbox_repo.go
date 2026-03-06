package postgres

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/user_service/internal/outbox"
)

type OutboxRepository struct {
	db *sql.DB
}

func (r *OutboxRepository) Insert(ctx context.Context, tx *sql.Tx, event *outbox.OutboxEvent) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO outbox_events 
    (id, aggregate_id, aggregate_type, event_type, payload, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		event.ID,
		event.AggregateID,
		event.AggregateType,
		event.EventType,
		event.Payload,
		event.Status,
		event.CreatedAt,
	)

	return err
}

func (r *OutboxRepository) FetchPending(ctx context.Context, limit int) ([]*outbox.OutboxEvent, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, aggregate_id, aggregate_type, event_type, payload, status, created_at
		 FROM outbox_events
		 WHERE status = 'pending'
		 ORDER BY created_at ASC
		 FOR UPDATE SKIP LOCKED
		 LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*outbox.OutboxEvent
	for rows.Next() {
		var e outbox.OutboxEvent
		err = rows.Scan(
			&e.ID,
			&e.AggregateID,
			&e.AggregateType,
			&e.EventType,
			&e.Payload,
			&e.Status,
			&e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, &e) // FIXME pointer or no?

	}
	return events, nil
}

func (r *OutboxRepository) MarkAsProcessed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE outbox_events
		 SET status = 'processed', processed_at = NOW()
		 WHERE id = $1`, id)

	return err
}
