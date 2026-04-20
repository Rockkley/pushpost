package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/database"
	apperr "github.com/rockkley/pushpost/services/message_service/internal/apperror"
	"github.com/rockkley/pushpost/services/message_service/internal/entity"
)

type MessageRepository struct {
	exec database.Executor
}

func NewMessageRepository(exec database.Executor) *MessageRepository {
	return &MessageRepository{exec: exec}
}

func (r *MessageRepository) Create(ctx context.Context, msg *entity.Message) (*entity.Message, error) {
	const query = `
		INSERT INTO messages (id, sender_id, receiver_id, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, sender_id, receiver_id, content, created_at, read_at`

	var out entity.Message
	err := r.exec.QueryRowContext(ctx, query,
		msg.ID, msg.SenderID, msg.ReceiverID, msg.Content, msg.CreatedAt,
	).Scan(&out.ID, &out.SenderID, &out.ReceiverID, &out.Content, &out.CreatedAt, &out.ReadAt)
	if err != nil {
		return nil, commonapperr.MapPostgresError(err, "create message")
	}
	return &out, nil
}

func (r *MessageRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Message, error) {
	const query = `
		SELECT id, sender_id, receiver_id, content, created_at, read_at
		FROM messages WHERE id = $1`

	var msg entity.Message
	err := r.exec.QueryRowContext(ctx, query, id).Scan(
		&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Content, &msg.CreatedAt, &msg.ReadAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperr.MessageNotFound()
		}
		return nil, commonapperr.MapPostgresError(err, "find message by id")
	}
	return &msg, nil
}

func (r *MessageRepository) GetConversation(
	ctx context.Context,
	userID, otherUserID uuid.UUID,
	limit, offset int,
) ([]*entity.Message, error) {
	const query = `
		SELECT id, sender_id, receiver_id, content, created_at, read_at
		FROM messages
		WHERE (sender_id = $1 AND receiver_id = $2)
		   OR (sender_id = $2 AND receiver_id = $1)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.exec.QueryContext(ctx, query, userID, otherUserID, limit, offset)
	if err != nil {
		return nil, commonapperr.MapPostgresError(err, "get conversation")
	}
	defer rows.Close()

	return scanMessages(rows)
}

func (r *MessageRepository) MarkAsRead(ctx context.Context, messageID, userID uuid.UUID) error {
	const query = `
		UPDATE messages
		SET read_at = $1
		WHERE id = $2 AND receiver_id = $3 AND read_at IS NULL`

	_, err := r.exec.ExecContext(ctx, query, time.Now().UTC(), messageID, userID)
	if err != nil {
		return commonapperr.MapPostgresError(err, "mark message as read")
	}
	return nil
}

func (r *MessageRepository) MarkAllAsRead(ctx context.Context, senderID, receiverID uuid.UUID) error {
	const query = `
		UPDATE messages
		SET read_at = $1
		WHERE sender_id = $2 AND receiver_id = $3 AND read_at IS NULL`

	_, err := r.exec.ExecContext(ctx, query, time.Now().UTC(), senderID, receiverID)
	if err != nil {
		return commonapperr.MapPostgresError(err, "mark all as read")
	}
	return nil
}

func (r *MessageRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM messages
		WHERE receiver_id = $1 AND read_at IS NULL`

	var count int
	err := r.exec.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, commonapperr.MapPostgresError(err, "get unread count")
	}
	return count, nil
}

func (r *MessageRepository) GetUnreadMessages(ctx context.Context, userID uuid.UUID) ([]*entity.Message, error) {
	const query = `
		SELECT id, sender_id, receiver_id, content, created_at, read_at
		FROM messages
		WHERE receiver_id = $1 AND read_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.exec.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, commonapperr.MapPostgresError(err, "get unread messages")
	}
	defer rows.Close()

	return scanMessages(rows)
}

func scanMessages(rows *sql.Rows) ([]*entity.Message, error) {
	var result []*entity.Message
	for rows.Next() {
		var msg entity.Message
		if err := rows.Scan(
			&msg.ID, &msg.SenderID, &msg.ReceiverID,
			&msg.Content, &msg.CreatedAt, &msg.ReadAt,
		); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		result = append(result, &msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return result, nil
}
