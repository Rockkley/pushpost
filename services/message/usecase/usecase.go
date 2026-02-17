package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/repository"
	"github.com/rockkley/pushpost/services/message/entity"
	"time"
)

type MessageService struct {
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
}

func NewMessageService(
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository) *MessageService {
	return &MessageService{messageRepo: messageRepo, userRepo: userRepo}
}

func (s *MessageService) SendMessage(ctx context.Context, senderID, receiverID uuid.UUID, content string) (*entity.Message, error) {
	if senderID == receiverID {
		return nil, entity.ErrCannotMessageSelf
	}

	if err := entity.ValidateContent(content); err != nil {
		return nil, err
	}

	// Checks that both users exists
	sender, err := s.userRepo.FindByID(ctx, senderID)
	if err != nil {
		return nil, err
	}

	if sender.IsDeleted() {
		return nil, entity.ErrSenderDeleted
	}

	receiver, err := s.userRepo.FindByID(ctx, receiverID)

	if err != nil {
		return nil, err
	}

	if receiver.IsDeleted() {
		return nil, entity.ErrReceiverDeleted
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
