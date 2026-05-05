package http

import (
	"encoding/json"
	commontransport "github.com/rockkley/pushpost/services/common_service/transport"
	"github.com/rockkley/pushpost/services/friendship_service/internal/entity"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/httperror"
	"github.com/rockkley/pushpost/services/friendship_service/internal/domain"
)

type FriendshipHandler struct {
	uc domain.FriendshipUseCase
}

func NewFriendshipHandler(uc domain.FriendshipUseCase) *FriendshipHandler {
	return &FriendshipHandler{uc: uc}
}

func (h *FriendshipHandler) SendRequest(w http.ResponseWriter, r *http.Request) error {
	senderID, err := commontransport.RequireUserID(r)

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
	slog.Debug("FriendshipHandler Accept Request")
	receiverID, err := commontransport.RequireUserID(r)

	if err != nil {

		return err
	}

	senderID, err := commontransport.ParsePathUUID(r, "senderID")

	if err != nil {

		return err
	}

	if err = h.uc.AcceptRequest(r.Context(), receiverID, senderID); err != nil {

		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "friend request accepted"})
}

func (h *FriendshipHandler) RejectRequest(w http.ResponseWriter, r *http.Request) error {
	receiverID, err := commontransport.RequireUserID(r)

	if err != nil {

		return err
	}

	senderID, err := commontransport.ParsePathUUID(r, "senderID")

	if err != nil {

		return err
	}

	if err = h.uc.RejectRequest(r.Context(), receiverID, senderID); err != nil {

		return err
	}
	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "friend request rejected"})
}

func (h *FriendshipHandler) CancelRequest(w http.ResponseWriter, r *http.Request) error {
	senderID, err := commontransport.RequireUserID(r)
	if err != nil {

		return err
	}
	receiverID, err := commontransport.ParsePathUUID(r, "receiverID")

	if err != nil {

		return err
	}

	if err = h.uc.CancelRequest(r.Context(), senderID, receiverID); err != nil {

		return err
	}
	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "friend request cancelled"})
}

func (h *FriendshipHandler) DeleteFriendship(w http.ResponseWriter, r *http.Request) error {
	userID, err := commontransport.RequireUserID(r)
	if err != nil {

		return err
	}
	friendID, err := commontransport.ParsePathUUID(r, "userID")
	if err != nil {

		return err
	}

	if err = h.uc.DeleteFriendship(r.Context(), userID, friendID); err != nil {

		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "friendship deleted"})
}

func (h *FriendshipHandler) GetFriendIDs(w http.ResponseWriter, r *http.Request) error {
	userID, err := commontransport.RequireUserID(r)

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
	userID, err := commontransport.RequireUserID(r)
	if err != nil {

		return err
	}

	friendID, err := commontransport.ParsePathUUID(r, "userID")

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
	viewerID, err := commontransport.RequireUserID(r)

	if err != nil {

		return err
	}
	targetID, err := commontransport.ParsePathUUID(r, "userID")
	if err != nil {

		return err
	}

	rel, err := h.uc.GetFriendshipStatus(r.Context(), viewerID, targetID)

	if err != nil {

		return err
	}

	status := &entity.FriendshipStatus{}
	if rel != nil {

		status = rel
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]bool{
		"are_friends":              status.AreFriends,
		"pending_request_sent":     status.PendingRequestSent,
		"pending_request_received": status.PendingRequestReceived,
	})
}

func (h *FriendshipHandler) GetIncomingRequests(w http.ResponseWriter, r *http.Request) error {
	receiverID, err := commontransport.RequireUserID(r)

	if err != nil {
		return err
	}

	reqs, err := h.uc.GetIncomingRequests(r.Context(), receiverID)

	if err != nil {
		return err
	}

	type item struct {
		RequestID string `json:"request_id"`
		SenderID  string `json:"sender_id"`
		CreatedAt string `json:"created_at"`
	}

	items := make([]item, 0, len(reqs))

	for _, req := range reqs {
		items = append(items, item{
			RequestID: req.ID.String(),
			SenderID:  req.SenderID.String(),
			CreatedAt: req.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]any{
		"requests": items,
		"count":    len(items),
	})
}

func (h *FriendshipHandler) GetOutgoingRequests(w http.ResponseWriter, r *http.Request) error {
	senderID, err := commontransport.RequireUserID(r)

	if err != nil {
		return err
	}

	reqs, err := h.uc.GetOutgoingRequests(r.Context(), senderID)

	if err != nil {
		return err
	}

	type item struct {
		RequestID  string `json:"request_id"`
		ReceiverID string `json:"receiver_id"`
		CreatedAt  string `json:"created_at"`
	}

	items := make([]item, 0, len(reqs))

	for _, req := range reqs {
		items = append(items, item{
			RequestID:  req.ID.String(),
			ReceiverID: req.ReceiverID.String(),
			CreatedAt:  req.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]any{
		"requests": items,
		"count":    len(items),
	})
}

func (h *FriendshipHandler) BlockUser(w http.ResponseWriter, r *http.Request) error {
	userID, err := commontransport.RequireUserID(r)
	if err != nil {
		return err
	}

	targetID, err := commontransport.ParsePathUUID(r, "userID")
	if err != nil {
		return err
	}

	if err = h.uc.BlockUser(r.Context(), userID, targetID); err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusCreated, map[string]string{"message": "user blocked"})
}

func (h *FriendshipHandler) UnblockUser(w http.ResponseWriter, r *http.Request) error {
	userID, err := commontransport.RequireUserID(r)
	if err != nil {
		return err
	}

	targetID, err := commontransport.ParsePathUUID(r, "userID")
	if err != nil {
		return err
	}

	if err = h.uc.UnblockUser(r.Context(), userID, targetID); err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "user unblocked"})
}

func (h *FriendshipHandler) AreBlocked(w http.ResponseWriter, r *http.Request) error {
	userID, err := commontransport.RequireUserID(r)
	if err != nil {
		return err
	}

	targetID, err := commontransport.ParsePathUUID(r, "userID")
	if err != nil {
		return err
	}

	blocked, err := h.uc.AreBlocked(r.Context(), userID, targetID)
	if err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]bool{"are_blocked": blocked})
}

func (h *FriendshipHandler) GetBlockedUsers(w http.ResponseWriter, r *http.Request) error {
	userID, err := commontransport.RequireUserID(r)
	if err != nil {
		return err
	}

	blockedIDs, err := h.uc.GetBlockedUserIDs(r.Context(), userID)
	if err != nil {
		return err
	}

	if blockedIDs == nil {
		blockedIDs = []uuid.UUID{}
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]any{
		"blocked_user_ids": blockedIDs,
		"count":            len(blockedIDs),
	})
}
