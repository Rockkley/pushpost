package postgtes

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	"time"
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

func (r *OutboxRepository) ClaimPending(ctx context.Context, limit int, maxAttempts int) ([]*outbox.OutboxEvent, error) {
	rows, err := r.db.QueryContext(ctx,
		`UPDATE outbox_events
	SET status = $1, updated_at = NOW()
	WHERE id IN (
    	SELECT id FROM outbox_events
              WHERE status = $2 
              AND attempts < $3
              ORDER BY created_at  
              LIMIT $4
              FOR UPDATE SKIP LOCKED
)
	RETURNING id, aggregate_id, aggregate_type, event_type, payload, attempts, created_at, updated_at`,
		outbox.StatusProcessing, outbox.StatusPending, maxAttempts, limit)
	if err != nil {
		return nil, fmt.Errorf("fetch pending events: %w", err)
	}
	defer rows.Close()

	var events []*outbox.OutboxEvent
	for rows.Next() {
		var e outbox.OutboxEvent
		if err = rows.Scan(
			&e.ID,
			&e.AggregateID,
			&e.AggregateType,
			&e.EventType,
			&e.Payload,
			&e.Status,
			&e.Attempts,
			&e.CreatedAt,
			&e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, &e)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("outbox rows: %w", err)
	}
	return events, nil
}

func (r *OutboxRepository) MarkAsProcessed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE outbox_events
		 SET status = $1, updated_at = NOW()
		 WHERE id = $2`, outbox.StatusProcessed, id)
	if err != nil {
		return fmt.Errorf("outbox mark processed: %w", err)
	}
	return nil

}

func (r *OutboxRepository) ResetStuck(ctx context.Context, stuckAfter time.Duration) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE outbox_events
    		 SET status = $1, 
    		     UPDATED_AT = NOW()
		 WHERE status = $2 AND updated_at < NOW() -($3 * INTERVAL '1 second')`,
		outbox.StatusPending, outbox.StatusProcessing, int(stuckAfter.Seconds()))
	if err != nil {
		return fmt.Errorf("reset stuck events: %w", err)
	}
	return nil
}

func (r *OutboxRepository) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE outbox_events
		 SET attempts = attempts + 1, 
		     status = $1,
		     updated_at = NOW()
		     WHERE id = $2`,
		outbox.StatusPending, id)

	if err != nil {

		return fmt.Errorf("increment attempts: %w", err)
	}

	return nil
}
