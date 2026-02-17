package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/message/entity"
)

type MessageRepository interface {
	Create(ctx context.Context, message *entity.Message) (*entity.Message, error)
	FindByUUID(ctx context.Context, id uuid.UUID) (*entity.Message, error)
	GetConversationWithUsers(ctx context.Context, userID, otherUserID uuid.UUID, limit, offset int) ([]*entity.Message, error)
	MarkAsRead(ctx context.Context, messageID, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, senderID, receiverID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	GetUnreadMessages(ctx context.Context, userID uuid.UUID) ([]*entity.Message, error)
}
