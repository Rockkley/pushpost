package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/httperror"
	"github.com/rockkley/pushpost/services/friendship_service/internal/domain"
	"github.com/rockkley/pushpost/services/friendship_service/internal/transport/http/middleware"
)

type FriendshipHandler struct {
	uc domain.FriendshipUseCase
}

func NewFriendshipHandler(uc domain.FriendshipUseCase) *FriendshipHandler {
	return &FriendshipHandler{uc: uc}
}

func (h *FriendshipHandler) SendRequest(w http.ResponseWriter, r *http.Request) error {
	senderID, err := requireUserID(r)
	if err != nil {
		return err
	}

	var body struct {
		ReceiverID uuid.UUID `json:"receiver_id"`
	}
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, "invalid JSON")
	}
	if body.ReceiverID == uuid.Nil {
		return commonapperr.Validation(commonapperr.CodeFieldRequired, "receiver_id", "receiver_id is required")
	}

	if err = h.uc.SendRequest(r.Context(), senderID, body.ReceiverID); err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusCreated, map[string]string{"message": "friend request sent"})
}

func (h *FriendshipHandler) AcceptRequest(w http.ResponseWriter, r *http.Request) error {
	receiverID, err := requireUserID(r)
	if err != nil {
		return err
	}
	senderID, err := parsePathUUID(r, "senderID")
	if err != nil {
		return err
	}

	if err = h.uc.AcceptRequest(r.Context(), receiverID, senderID); err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "friend request accepted"})
}

func (h *FriendshipHandler) RejectRequest(w http.ResponseWriter, r *http.Request) error {
	receiverID, err := requireUserID(r)
	if err != nil {
		return err
	}
	senderID, err := parsePathUUID(r, "senderID")
	if err != nil {
		return err
	}

	if err = h.uc.RejectRequest(r.Context(), receiverID, senderID); err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "friend request rejected"})
}

func (h *FriendshipHandler) CancelRequest(w http.ResponseWriter, r *http.Request) error {
	senderID, err := requireUserID(r)
	if err != nil {
		return err
	}
	receiverID, err := parsePathUUID(r, "receiverID")
	if err != nil {
		return err
	}

	if err = h.uc.CancelRequest(r.Context(), senderID, receiverID); err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "friend request cancelled"})
}

func (h *FriendshipHandler) DeleteFriendship(w http.ResponseWriter, r *http.Request) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}
	friendID, err := parsePathUUID(r, "userID")
	if err != nil {
		return err
	}

	if err = h.uc.DeleteFriendship(r.Context(), userID, friendID); err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "friendship deleted"})
}

func (h *FriendshipHandler) GetFriendIDs(w http.ResponseWriter, r *http.Request) error {
	userID, err := requireUserID(r)

	if err != nil {

		return err
	}

	ids, err := h.uc.GetFriendIDs(r.Context(), userID)

	if err != nil {

		return err
	}

	if ids == nil {
		ids = []uuid.UUID{}
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]any{
		"friend_ids": ids,
		"count":      len(ids),
	})
}

func (h *FriendshipHandler) AreFriends(w http.ResponseWriter, r *http.Request) error {
	userID, err := requireUserID(r)
	if err != nil {

		return err
	}

	friendID, err := parsePathUUID(r, "userID")

	if err != nil {

		return err
	}

	ok, err := h.uc.AreFriends(r.Context(), userID, friendID)

	if err != nil {

		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]bool{"are_friends": ok})
}

func (h *FriendshipHandler) GetRelationship(w http.ResponseWriter, r *http.Request) error {
	viewerID, err := requireUserID(r)
	if err != nil {
		return err
	}
	targetID, err := parsePathUUID(r, "userID")
	if err != nil {
		return err
	}

	rel, err := h.uc.GetFriendshipStatus(r.Context(), viewerID, targetID)
	if err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]bool{
		"are_friends":              rel.AreFriends,
		"pending_request_sent":     rel.PendingRequestSent,
		"pending_request_received": rel.PendingRequestReceived,
	})
}

func requireUserID(r *http.Request) (uuid.UUID, error) {
	userID, ok := r.Context().Value(middleware.CtxUserIDKey).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return uuid.Nil, commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "missing authenticated user")
	}
	return userID, nil
}

func parsePathUUID(r *http.Request, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, param))
	if err != nil {
		return uuid.Nil, commonapperr.BadRequest(
			commonapperr.CodeFieldInvalid, "invalid "+param+" — must be a UUID",
		)
	}
	return id, nil
}
