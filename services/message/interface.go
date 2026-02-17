package message

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/message/entity"
)

type MessageService interface {
	SendMessage(ctx context.Context, senderID, receiverID uuid.UUID, content string) (*entity.Message, error)
}
