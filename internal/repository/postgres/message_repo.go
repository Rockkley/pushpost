package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/domain"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, msg *domain.Message) (*domain.Message, error) {
	query := `
		INSERT INTO messages (id, sender_id, receiver_id, content, created_at)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, sender_id, receiver_id, content, created_at, read_at `

	var newMsg domain.Message
	err := r.db.QueryRowContext(ctx, query,
		msg.Id, msg.SenderID, msg.ReceiverID, msg.Content, msg.CreatedAt,
	).Scan(
		&newMsg.Id,
		&newMsg.SenderID,
		&newMsg.ReceiverID,
		&newMsg.Content,
		&newMsg.CreatedAt,
		&newMsg.ReadAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return &newMsg, nil
}

func (r *MessageRepository) FindByUUID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	query := `
		SELECT id, sender_id, receiver_id, content, created_at, read_at
		FROM messages
		WHERE id=$1`

	var msg domain.Message

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.Id,
		&msg.SenderID,
		&msg.ReceiverID,
		&msg.Content,
		&msg.CreatedAt,
		&msg.ReadAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrMessageNotFound
		}

		return nil, fmt.Errorf("failed to find message by uuid: %w", err)
	}

	return &msg, nil
}
