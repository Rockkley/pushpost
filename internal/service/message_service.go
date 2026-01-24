package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/domain"
	"github.com/rockkley/pushpost/internal/repository"
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

func (s *MessageService) SendMessage(ctx context.Context, senderID, receiverID uuid.UUID, content string) (*domain.Message, error) {
	if senderID == receiverID {
		return nil, domain.ErrCannotMessageSelf
	}

	if err := domain.ValidateContent(content); err != nil {
		return nil, err
	}

	// Checks that both users exists
	sender, err := s.userRepo.FindByID(ctx, senderID)
	if err != nil {
		return nil, err
	}

	if sender.IsDeleted() {
		return nil, domain.ErrSenderDeleted
	}

	receiver, err := s.userRepo.FindByID(ctx, receiverID)

	if err != nil {
		return nil, err
	}

	if receiver.IsDeleted() {
		return nil, domain.ErrReceiverDeleted
	}

	msg := &domain.Message{
		Id:         uuid.New(),
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		CreatedAt:  time.Now(),
	}
	return s.messageRepo.Create(ctx, msg)

}
