package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	"github.com/rockkley/pushpost/services/message_service/internal/domain/dto"
	"github.com/rockkley/pushpost/services/message_service/internal/entity"
	"github.com/rockkley/pushpost/services/message_service/internal/repository"
)

type MessageUseCase interface {
	SendMessage(ctx context.Context, req dto.SendMessageDTO) (*entity.Message, error)
	GetConversation(ctx context.Context, req dto.GetConversationDTO) ([]*entity.Message, error)
	MarkAsRead(ctx context.Context, messageID, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, senderID, receiverID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	GetUnreadMessages(ctx context.Context, userID uuid.UUID) ([]*entity.Message, error)
}

type Tx interface {
	Messages() repository.MessageRepository
	Outbox() outbox.WriterInterface
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(tx Tx) error) error
	Reader() repository.MessageRepository
}
