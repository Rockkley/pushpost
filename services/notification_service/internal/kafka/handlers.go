package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/notification_service/internal/domain"
	"github.com/rockkley/pushpost/services/notification_service/internal/entity"
)

type Handlers struct {
	uc  domain.NotificationUseCase
	log *slog.Logger
}

func NewHandlers(uc domain.NotificationUseCase, log *slog.Logger) *Handlers {
	return &Handlers{uc: uc, log: log.With("component", "notification_handlers")}
}

func (h *Handlers) HandleFriendRequestSent(ctx context.Context, payload json.RawMessage) error {
	var p domain.FriendRequestSentPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("decode friend_request.sent: %w", err)
	}
	receiverID, err := uuid.Parse(p.ReceiverID)
	if err != nil {
		h.log.Warn("invalid receiver_id in friend_request.sent, skipping", slog.String("receiver_id", p.ReceiverID))
		return nil
	}
	return h.uc.CreateAndDeliver(ctx, &entity.Notification{ID: uuid.New(), UserID: receiverID, Type: entity.TypeFriendRequestReceived, Title: "Новая заявка в друзья", Body: "Кто-то хочет добавить вас в друзья.", Data: map[string]string{"request_id": p.RequestID, "sender_id": p.SenderID}})
}

func (h *Handlers) HandleFriendshipCreated(ctx context.Context, payload json.RawMessage) error {
	var p domain.FriendshipCreatedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("decode friendship.created: %w", err)
	}
	user1ID, err := uuid.Parse(p.User1ID)
	if err != nil {
		h.log.Warn("invalid user1_id in friendship.created, skipping", slog.String("user1_id", p.User1ID))
		return nil
	}
	user2ID, err := uuid.Parse(p.User2ID)
	if err != nil {
		h.log.Warn("invalid user2_id in friendship.created, skipping", slog.String("user2_id", p.User2ID))
		return nil
	}
	for _, userID := range []uuid.UUID{user1ID, user2ID} {
		if err = h.uc.CreateAndDeliver(ctx, &entity.Notification{ID: uuid.New(), UserID: userID, Type: entity.TypeFriendRequestAccepted, Title: "Заявка принята", Body: "Вы теперь друзья!", Data: map[string]string{"friendship_id": p.FriendshipID}}); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handlers) HandleFriendRequestRejected(ctx context.Context, payload json.RawMessage) error {
	var p domain.FriendRequestRejectedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("decode friend_request.rejected: %w", err)
	}
	senderID, err := uuid.Parse(p.SenderID)
	if err != nil {
		h.log.Warn("invalid sender_id in friend_request.rejected, skipping", slog.String("sender_id", p.SenderID))
		return nil
	}
	return h.uc.CreateAndDeliver(ctx, &entity.Notification{ID: uuid.New(), UserID: senderID, Type: entity.TypeFriendRequestRejected, Title: "Заявка отклонена", Body: "Пользователь отклонил вашу заявку в друзья.", Data: map[string]string{"receiver_id": p.ReceiverID}})
}

func (h *Handlers) HandleMessageSent(ctx context.Context, payload json.RawMessage) error {
	var p domain.MessageSentPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("decode message.sent: %w", err)
	}
	receiverID, err := uuid.Parse(p.ReceiverID)
	if err != nil {
		h.log.Warn("invalid receiver_id in message.sent, skipping", slog.String("receiver_id", p.ReceiverID))
		return nil
	}
	return h.uc.CreateAndDeliver(ctx, &entity.Notification{ID: uuid.New(), UserID: receiverID, Type: entity.TypeMessageReceived, Title: "Новое сообщение", Body: "У вас новое личное сообщение.", Data: map[string]string{"message_id": p.MessageID, "sender_id": p.SenderID}})
}
