package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	apperr "github.com/rockkley/pushpost/services/message_service/internal/apperror"
	"github.com/rockkley/pushpost/services/message_service/internal/domain"
	"github.com/rockkley/pushpost/services/message_service/internal/domain/dto"
	"github.com/rockkley/pushpost/services/message_service/internal/entity"
	"log/slog"
)

const (
	defaultConversationLimit = 50
	maxConversationLimit     = 200
)

type MessageUseCase struct {
	uow domain.UnitOfWork
}

func NewMessageUseCase(uow domain.UnitOfWork) *MessageUseCase {
	return &MessageUseCase{uow: uow}
}

func (uc *MessageUseCase) SendMessage(ctx context.Context, req dto.SendMessageDTO) (*entity.Message, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "MessageUseCase.SendMessage"))

	if req.SenderID == req.ReceiverID {

		return nil, apperr.CannotMessageSelf()
	}

	if err := entity.ValidateContent(req.Content); err != nil {
		switch err {
		case entity.ErrMessageEmpty:
			return nil, apperr.MessageEmpty()
		case entity.ErrMessageTooLong:
			return nil, apperr.MessageTooLong()
		default:
			return nil, apperr.MessageEmpty()
		}
	}

	msg := &entity.Message{
		ID:         uuid.New(),
		SenderID:   req.SenderID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
		CreatedAt:  time.Now().UTC(),
	}

	var created *entity.Message

	err := uc.uow.Do(ctx, func(tx domain.Tx) error {
		var txErr error
		created, txErr = tx.Messages().Create(ctx, msg)

		if txErr != nil {

			return txErr
		}

		return insertOutboxEvent(ctx, tx, created)
	})

	if err != nil {
		log.Error("failed to send message", slog.Any("error", err))

		return nil, err
	}

	log.Info("message sent",
		slog.String("message_id", created.ID.String()),
		slog.String("sender_id", req.SenderID.String()),
		slog.String("receiver_id", req.ReceiverID.String()),
	)

	return created, nil
}

func (uc *MessageUseCase) GetConversation(ctx context.Context, req dto.GetConversationDTO) ([]*entity.Message, error) {
	limit := req.Limit

	if limit <= 0 {
		limit = defaultConversationLimit
	}

	if limit > maxConversationLimit {
		limit = maxConversationLimit
	}

	offset := req.Offset

	if offset < 0 {
		offset = 0
	}

	return uc.uow.Reader().GetConversation(ctx, req.UserID, req.OtherUserID, limit, offset)
}

func (uc *MessageUseCase) MarkAsRead(ctx context.Context, messageID, userID uuid.UUID) error {
	log := ctxlog.From(ctx).With(slog.String("op", "MessageUseCase.MarkAsRead"))

	msg, err := uc.uow.Reader().FindByID(ctx, messageID)
	if err != nil {

		return err
	}

	if msg.ReceiverID != userID {

		return apperr.NotReceiver()
	}

	if err = uc.uow.Reader().MarkAsRead(ctx, messageID, userID); err != nil {

		return err
	}

	log.Info("message marked as read", slog.String("message_id", messageID.String()))

	return nil
}

func (uc *MessageUseCase) MarkAllAsRead(ctx context.Context, senderID, receiverID uuid.UUID) error {
	log := ctxlog.From(ctx).With(slog.String("op", "MessageUseCase.MarkAllAsRead"))

	if err := uc.uow.Reader().MarkAllAsRead(ctx, senderID, receiverID); err != nil {

		return err
	}

	log.Info("all messages marked as read",
		slog.String("sender_id", senderID.String()),
		slog.String("receiver_id", receiverID.String()),
	)

	return nil
}

func (uc *MessageUseCase) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return uc.uow.Reader().GetUnreadCount(ctx, userID)
}

func (uc *MessageUseCase) GetUnreadMessages(ctx context.Context, userID uuid.UUID) ([]*entity.Message, error) {
	return uc.uow.Reader().GetUnreadMessages(ctx, userID)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func insertOutboxEvent(ctx context.Context, tx domain.Tx, msg *entity.Message) error {
	eventPayload := entity.MessageSentEvent{
		MessageID:  msg.ID.String(),
		SenderID:   msg.SenderID.String(),
		ReceiverID: msg.ReceiverID.String(),
		CreatedAt:  msg.CreatedAt.Format(time.RFC3339),
	}

	inner, err := json.Marshal(eventPayload)

	if err != nil {

		return commonapperr.Internal("marshal message.sent payload", err)
	}

	type envelope struct {
		EventType string          `json:"event_type"`
		Payload   json.RawMessage `json:"payload"`
	}

	payload, err := json.Marshal(envelope{
		EventType: entity.EventMessageSent,
		Payload:   inner,
	})

	if err != nil {

		return commonapperr.Internal("marshal message.sent envelope", err)
	}

	return tx.Outbox().Insert(ctx, &outbox.OutboxEvent{
		ID:            uuid.New(),
		AggregateID:   msg.ID.String(),
		AggregateType: "message",
		EventType:     entity.EventMessageSent,
		Payload:       payload,
	})
}
