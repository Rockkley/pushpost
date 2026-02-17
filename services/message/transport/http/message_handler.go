package http

import (
	"github.com/rockkley/pushpost/internal/services/message/usecase"
)

type MessageHandler struct {
	messageUsecase usecase.MessageUsecase
}

func NewMessageHandler(messageUsecase services.messageUsecase) *MessageHandler {
	return &MessageHandler{messageUsecase: messageUsecase}
}
