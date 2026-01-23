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

func (r *MessageRepository) Create(ctx context.Context, senderID, receiverID uuid.UUID, content string) (*domain.Message, error) {
	query := `
		INSERT INTO messages (sender_id, receiver_id, content)
		VALUES ($1,$2,$3)
		RETURNING id, sender_id, receiver_id, content, created_at, read_at`

	var msg domain.Message
	err := r.db.QueryRowContext(ctx, query, senderID, receiverID, content).Scan(
		&msg.ID,
		&msg.SenderID,
		&msg.ReceiverID,
		&msg.Content,
		&msg.CreatedAt,
		&msg.ReadAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return &msg, nil
}

func (r *MessageRepository) FindByUUID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	query := `
		SELECT id, sender_id, receiver_id, content, created_at, read_at
		FROM messages
		WHERE id=$1`

	var msg domain.Message

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID,
		&msg.SenderID,
		&msg.ReceiverID,
		&msg.Content,
		&msg.CreatedAt,
		&msg.ReadAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrMessageNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find message by uuid: %w", err)
	}

	return &msg, nil
}
