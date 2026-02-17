package postrgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/message/entity"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, msg *entity.Message) (*entity.Message, error) {
	query := `
		INSERT INTO messages (id, sender_id, receiver_id, content, created_at)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, sender_id, receiver_id, content, created_at, read_at `

	var newMsg entity.Message
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

func (r *MessageRepository) FindByUUID(ctx context.Context, id uuid.UUID) (*entity.Message, error) {
	query := `
		SELECT id, sender_id, receiver_id, content, created_at, read_at
		FROM messages
		WHERE id=$1`

	var msg entity.Message

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
			return nil, entity.ErrMessageNotFound
		}

		return nil, fmt.Errorf("failed to find message by uuid: %w", err)
	}

	return &msg, nil
}

func (r *MessageRepository) GetConversationWithUsers(ctx context.Context, userID, otherUserID uuid.UUID, limit, offset int) ([]*entity.Message, error) {
	//TODO implement me
	panic("implement me")
}

func (r *MessageRepository) MarkAsRead(ctx context.Context, messageID, userID uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (r *MessageRepository) MarkAllAsRead(ctx context.Context, senderID, receiverID uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (r *MessageRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (r *MessageRepository) GetUnreadMessages(ctx context.Context, userID uuid.UUID) ([]*entity.Message, error) {
	//TODO implement me
	panic("implement me")
}
