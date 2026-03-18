package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/message_service/internal/entity"
	"github.com/rockkley/pushpost/services/message_service/internal/repository"
	"time"
)

type MessageService struct {
	messageRepo repository.MessageRepository
}

func NewMessageService(
	messageRepo repository.MessageRepository) *MessageService {
	return &MessageService{messageRepo: messageRepo}
}

func (s *MessageService) SendMessage(ctx context.Context, senderID, receiverID uuid.UUID, content string) (*entity.Message, error) {
	if senderID == receiverID {
		return nil, entity.ErrCannotMessageSelf
	}

	if err := entity.ValidateContent(content); err != nil {
		return nil, err
	}

	msg := &entity.Message{
		Id:         uuid.New(),
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		CreatedAt:  time.Now(),
	}
	return s.messageRepo.Create(ctx, msg)

}
