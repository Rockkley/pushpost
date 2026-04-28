package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/httperror"
	commonmiddleware "github.com/rockkley/pushpost/services/common_service/middleware"
	"github.com/rockkley/pushpost/services/notification_service/internal/domain"
	"github.com/rockkley/pushpost/services/notification_service/internal/entity"
)

type NotificationHandler struct{ uc domain.NotificationUseCase }

func NewNotificationHandler(uc domain.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{uc: uc}
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	notifications, err := h.uc.GetForUser(r.Context(), userID, limit, offset)
	if err != nil {
		return err
	}
	if notifications == nil {
		notifications = []*entity.Notification{}
	}
	return httperror.WriteJSON(w, http.StatusOK, map[string]any{
		"notifications": notifications,
		"count":         len(notifications),
	})
}

func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}
	count, err := h.uc.GetUnreadCount(r.Context(), userID)
	if err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusOK, map[string]int{"unread_count": count})
}

func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}
	notificationID, err := uuid.Parse(chi.URLParam(r, "notificationID"))
	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid notification id")
	}
	if err = h.uc.MarkAsRead(r.Context(), notificationID, userID); err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "marked as read"})
}

func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}
	if err = h.uc.MarkAllAsRead(r.Context(), userID); err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "all marked as read"})
}

func (h *NotificationHandler) GetPreferences(w http.ResponseWriter, r *http.Request) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}
	prefs, err := h.uc.GetPreferences(r.Context(), userID)
	if err != nil {
		return err
	}
	if prefs == nil {
		prefs = []*entity.NotificationPreference{}
	}
	return httperror.WriteJSON(w, http.StatusOK, map[string]any{"preferences": prefs})
}

func (h *NotificationHandler) SetPreference(w http.ResponseWriter, r *http.Request) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}

	var body struct {
		Type    string `json:"type"`
		Channel string `json:"channel"`
		Enabled bool   `json:"enabled"`
	}
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, "invalid JSON")
	}

	if body.Type == "" {
		return commonapperr.Validation(commonapperr.CodeFieldRequired, "type", "type is required")
	}
	if body.Channel == "" {
		return commonapperr.Validation(commonapperr.CodeFieldRequired, "channel", "channel is required")
	}

	// Проверяем допустимые значения channel.
	// Недопустимые строки не должны попадать в БД:
	// они нарушат CHECK constraint и приведут к молчаливой потере доставки.
	if !isValidChannel(entity.Channel(body.Channel)) {
		return commonapperr.Validation(
			commonapperr.CodeFieldInvalid, "channel",
			"unknown channel, allowed: in_app, telegram",
		)
	}

	// Проверяем допустимые значения type.
	if !isValidNotificationType(entity.NotificationType(body.Type)) {
		return commonapperr.Validation(
			commonapperr.CodeFieldInvalid, "type",
			"unknown notification type",
		)
	}

	if err = h.uc.SetPreference(r.Context(), &entity.NotificationPreference{
		UserID:  userID,
		Type:    entity.NotificationType(body.Type),
		Channel: entity.Channel(body.Channel),
		Enabled: body.Enabled,
	}); err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "preference updated"})
}

func (h *NotificationHandler) GenerateTelegramCode(w http.ResponseWriter, r *http.Request) error {
	userID, err := requireUserID(r)

	if err != nil {
		return err
	}

	code, err := h.uc.GenerateTelegramLinkCode(r.Context(), userID)

	if err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"code": code})
}

func (h *NotificationHandler) UnbindTelegram(w http.ResponseWriter, r *http.Request) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}
	if err = h.uc.UnbindTelegram(r.Context(), userID); err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "telegram unbound"})
}

// ── helpers ───────────────────────────────────────────────────────────────────

func requireUserID(r *http.Request) (uuid.UUID, error) {
	userID, ok := commonmiddleware.UserIDFromContext(r.Context())
	if !ok || userID == uuid.Nil {
		return uuid.Nil, commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "missing authenticated user")
	}
	return userID, nil
}

// isValidChannel проверяет, что channel входит в допустимое множество.
// Является единственным источником правды на уровне транспорта.
// Допустимые значения должны совпадать с entity.AllChannels и CHECK constraint в миграции.
func isValidChannel(ch entity.Channel) bool {
	for _, allowed := range entity.AllChannels {
		if ch == allowed {
			return true
		}
	}
	return false
}

// isValidNotificationType проверяет, что type входит в допустимое множество.
func isValidNotificationType(t entity.NotificationType) bool {
	switch t {
	case entity.TypeFriendRequestReceived,
		entity.TypeFriendRequestAccepted,
		entity.TypeFriendRequestRejected,
		entity.TypeMessageReceived:
		return true
	}
	return false
}
