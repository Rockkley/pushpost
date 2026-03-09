package http

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/service"
	"github.com/rockkley/pushpost/services/common_service/handler/http/middleware"
	"github.com/rockkley/pushpost/services/common_service/httperror"
	"github.com/rockkley/pushpost/services/friendship_service/internal/transport/http/dto"
	"net/http"
)

type FriendHandler struct {
	friendService service.FriendService
}

func (h *FriendHandler) SendRequest(w http.ResponseWriter, r *http.Request) error {
	userID := r.Context().Value(middleware.CtxUserIDKey).(uuid.UUID)

	var req dto.SendRequestDTO

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {

		return err
	}

	req.SenderID = userID

	if err = h.friendService.SendRequest(r.Context(), req); err != nil {

		return err
	}

	return httperror.WriteJSON(w, http.StatusCreated, map[string]string{
		"message_service": "friend request sent",
	})
}
