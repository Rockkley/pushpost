package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/clients/profile_grpc"
	"github.com/rockkley/pushpost/services/notification_service/internal/domain"
	"github.com/rockkley/pushpost/services/notification_service/internal/entity"
)

type Handlers struct {
	uc            domain.NotificationUseCase
	log           *slog.Logger
	profileClient *profile_grpc.Client
}

func NewHandlers(uc domain.NotificationUseCase, profileClient *profile_grpc.Client, log *slog.Logger) *Handlers {
	return &Handlers{uc: uc, profileClient: profileClient, log: log.With("component", "notification_handlers")}
}

func notifID(key string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(key))
}

func (h *Handlers) HandleFriendRequestSent(ctx context.Context, payload json.RawMessage) error {
	var p domain.FriendRequestSentPayload

	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("decode friend_request.sent: %w", err)
	}

	receiverID, err := uuid.Parse(p.ReceiverID)

	if err != nil {
		h.log.Warn("invalid receiver_id in friend_request.sent, skipping",
			slog.String("receiver_id", p.ReceiverID))
		return nil
	}

	return h.uc.CreateAndDeliver(ctx, &entity.Notification{
		ID:     notifID("friend_request.received:" + p.RequestID),
		UserID: receiverID,
		Type:   entity.TypeFriendRequestReceived,
		Title:  "Новая заявка в друзья",
		Body:   "Кто-то хочет добавить вас в друзья.",
		Data:   map[string]string{"request_id": p.RequestID, "sender_id": p.SenderID},
	})
}

func (h *Handlers) HandleFriendshipCreated(ctx context.Context, payload json.RawMessage) error {
	var p domain.FriendshipCreatedPayload

	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("decode friendship.created: %w", err)
	}

	user1ID, err := uuid.Parse(p.User1ID)
	if err != nil {
		h.log.Warn("invalid user1_id in friendship.created, skipping",
			slog.String("user1_id", p.User1ID))
		return nil
	}

	user2ID, err := uuid.Parse(p.User2ID)
	if err != nil {
		h.log.Warn("invalid user2_id in friendship.created, skipping",
			slog.String("user2_id", p.User2ID))
		return nil
	}

	notif1 := &entity.Notification{
		ID:     notifID("friendship.created:" + p.FriendshipID + ":user1"),
		UserID: user1ID,
		Type:   entity.TypeFriendRequestAccepted,
		Title:  "Заявка принята",
		Body:   "Вы теперь друзья!",
		Data:   map[string]string{"friendship_id": p.FriendshipID},
	}

	notif2 := &entity.Notification{
		ID:     notifID("friendship.created:" + p.FriendshipID + ":user2"),
		UserID: user2ID,
		Type:   entity.TypeFriendRequestAccepted,
		Title:  "Заявка принята",
		Body:   "Вы теперь друзья!",
		Data:   map[string]string{"friendship_id": p.FriendshipID},
	}

	if err = h.uc.CreateAndDeliver(ctx, notif1); err != nil {
		return err
	}

	return h.uc.CreateAndDeliver(ctx, notif2)
}

func (h *Handlers) HandleFriendRequestRejected(ctx context.Context, payload json.RawMessage) error {
	var p domain.FriendRequestRejectedPayload

	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("decode friend_request.rejected: %w", err)
	}

	senderID, err := uuid.Parse(p.SenderID)
	if err != nil {
		h.log.Warn("invalid sender_id in friend_request.rejected, skipping",
			slog.String("sender_id", p.SenderID))
		return nil
	}

	return h.uc.CreateAndDeliver(ctx, &entity.Notification{
		ID:     notifID("friend_request.rejected:" + p.SenderID + ":" + p.ReceiverID),
		UserID: senderID,
		Type:   entity.TypeFriendRequestRejected,
		Title:  "Заявка отклонена",
		Body:   "Пользователь отклонил вашу заявку в друзья.",
		Data:   map[string]string{"receiver_id": p.ReceiverID},
	})
}

func (h *Handlers) HandleMessageSent(ctx context.Context, payload json.RawMessage) error {
	var p domain.MessageSentPayload

	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("decode message.sent: %w", err)
	}

	receiverID, err := uuid.Parse(p.ReceiverID)
	if err != nil {
		h.log.Warn("invalid receiver_id in message.sent, skipping",
			slog.String("receiver_id", p.ReceiverID))
		return nil
	}

	return h.uc.CreateAndDeliver(ctx, &entity.Notification{
		ID:     notifID("message.received:" + p.MessageID),
		UserID: receiverID,
		Type:   entity.TypeMessageReceived,
		Title:  "Новое сообщение",
		Body:   "У вас новое личное сообщение.",
		Data:   map[string]string{"message_id": p.MessageID, "sender_id": p.SenderID},
	})
}

func (h *Handlers) HandleCommentReplied(ctx context.Context, payload json.RawMessage) error {
	var p domain.CommentRepliedPayload

	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("decode comment.replied: %w", err)
	}

	userID, err := uuid.Parse(p.OriginalAuthorID)

	if err != nil {
		return nil
	}

	if p.ReplyAuthorID == p.OriginalAuthorID {
		return nil
	}
	return h.uc.CreateAndDeliver(ctx, &entity.Notification{
		ID:     notifID("comment.replied:" + p.CommentID + ":" + p.OriginalAuthorID),
		UserID: userID,
		Type:   entity.TypeCommentReplied,
		Title:  "Новый ответ на комментарий",
		Body:   "Кто-то ответил на ваш комментарий.",
		Data:   map[string]string{"post_id": p.PostID, "comment_id": p.CommentID, "parent_comment_id": p.ParentCommentID, "reply_author_id": p.ReplyAuthorID},
	})
}

func (h *Handlers) HandleCommentMentioned(ctx context.Context, payload json.RawMessage) error {
	var p domain.CommentMentionedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("decode comment.mentioned: %w", err)
	}
	for _, username := range p.MentionedList {
		uname := strings.TrimSpace(strings.ToLower(username))
		if uname == "" || h.profileClient == nil {
			continue
		}
		profile, err := h.profileClient.GetByUsername(ctx, uname)
		if err != nil {
			continue
		}
		userID, err := uuid.Parse(profile.UserID)
		if err != nil || userID.String() == p.AuthorID {
			continue
		}
		_ = h.uc.CreateAndDeliver(ctx, &entity.Notification{
			ID:     notifID("comment.mentioned:" + p.CommentID + ":" + userID.String()),
			UserID: userID,
			Type:   entity.TypeCommentMentioned,
			Title:  "Вас упомянули в комментарии",
			Body:   "Кто-то упомянул вас в комментарии.",
			Data:   map[string]string{"post_id": p.PostID, "comment_id": p.CommentID, "author_id": p.AuthorID},
		})
	}
	return nil
}
