package http

import (
	"encoding/json"
	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/httperror"
	commontransport "github.com/rockkley/pushpost/services/common_service/transport"
	"github.com/rockkley/pushpost/services/message_service/internal/domain"
	"github.com/rockkley/pushpost/services/message_service/internal/domain/dto"
	"github.com/rockkley/pushpost/services/message_service/internal/entity"
	"net/http"
)

type MessageHandler struct {
	uc domain.MessageUseCase
}

func NewMessageHandler(uc domain.MessageUseCase) *MessageHandler {
	return &MessageHandler{uc: uc}
}

func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) error {
	senderID, err := commontransport.RequireUserID(r)

	if err != nil {

		return err
	}

	var body struct {
		ReceiverID uuid.UUID `json:"receiver_id"`
		Content    string    `json:"content"`
	}
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {

		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, "invalid JSON")
	}

	if body.ReceiverID == uuid.Nil {

		return commonapperr.Validation(commonapperr.CodeFieldRequired, "receiver_id", "receiver_id is required")
	}

	if body.Content == "" {

		return commonapperr.Validation(commonapperr.CodeFieldRequired, "content", "content is required")
	}

	msg, err := h.uc.SendMessage(r.Context(), dto.SendMessageDTO{
		SenderID:   senderID,
		ReceiverID: body.ReceiverID,
		Content:    body.Content,
	})

	if err != nil {

		return err
	}

	return httperror.WriteJSON(w, http.StatusCreated, msg)
}

func (h *MessageHandler) GetConversation(w http.ResponseWriter, r *http.Request) error {
	userID, err := commontransport.RequireUserID(r)

	if err != nil {

		return err
	}

	otherUserID, err := commontransport.ParsePathUUID(r, "userID")

	if err != nil {

		return err
	}

	limit, offset, err := commontransport.ParsePagination(r)

	if err != nil {

		return err
	}

	messages, err := h.uc.GetConversation(r.Context(), dto.GetConversationDTO{
		UserID:      userID,
		OtherUserID: otherUserID,
		Limit:       limit,
		Offset:      offset,
	})

	if err != nil {

		return err
	}

	if messages == nil {
		messages = []*entity.Message{}
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]any{
		"messages": messages,
		"count":    len(messages),
	})
}

func (h *MessageHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) error {
	userID, err := commontransport.RequireUserID(r)

	if err != nil {

		return err
	}

	messageID, err := commontransport.ParsePathUUID(r, "messageID")

	if err != nil {

		return err
	}

	if err = h.uc.MarkAsRead(r.Context(), messageID, userID); err != nil {

		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "marked as read"})
}

func (h *MessageHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) error {
	receiverID, err := commontransport.RequireUserID(r)

	if err != nil {

		return err
	}

	senderID, err := commontransport.ParsePathUUID(r, "userID")

	if err != nil {

		return err
	}

	if err = h.uc.MarkAllAsRead(r.Context(), senderID, receiverID); err != nil {

		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "all messages marked as read"})
}

func (h *MessageHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) error {
	userID, err := commontransport.RequireUserID(r)

	if err != nil {

		return err
	}

	count, err := h.uc.GetUnreadCount(r.Context(), userID)

	if err != nil {

		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]int{"unread_count": count})
}

func (h *MessageHandler) GetUnreadMessages(w http.ResponseWriter, r *http.Request) error {
	userID, err := commontransport.RequireUserID(r)

	if err != nil {

		return err
	}

	messages, err := h.uc.GetUnreadMessages(r.Context(), userID)

	if err != nil {

		return err
	}

	if messages == nil {
		messages = []*entity.Message{}
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]any{
		"messages": messages,
		"count":    len(messages),
	})
}
