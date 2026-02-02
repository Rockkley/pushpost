package http

import "github.com/rockkley/pushpost/internal/service"

type MessageHandler struct {
	messageService service.MessageService
}

func NewMessageHandler(messageService service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}
