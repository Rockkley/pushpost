package services

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	apperr "github.com/rockkley/pushpost/services/friendship_service/internal/apperror"
	"github.com/rockkley/pushpost/services/friendship_service/internal/domain"
	"github.com/rockkley/pushpost/services/friendship_service/internal/entity"
	"log/slog"
)

type FriendshipUseCase struct {
	uow domain.UnitOfWork
}

func NewFriendshipUseCase(uow domain.UnitOfWork) *FriendshipUseCase {
	return &FriendshipUseCase{uow: uow}
}

func (uc *FriendshipUseCase) SendRequest(ctx context.Context, senderID, receiverID uuid.UUID) error {
	if senderID == receiverID {

		return apperr.CannotBefriendSelf()
	}

	err := uc.uow.Do(ctx, func(tx domain.Tx) error {
		alreadyFriends, err := tx.Friendships().Exists(ctx, senderID, receiverID)

		if err != nil {

			return err
		}

		if alreadyFriends {

			return apperr.AlreadyFriends()
		}

		reqExists, err := tx.Requests().FindPendingBetween(ctx, senderID, receiverID)

		if err != nil {

			return err
		}

		if reqExists != nil {

			return apperr.FriendRequestExists()
		}

		req := entity.FriendshipRequest{
			ID:         uuid.New(),
			SenderID:   senderID,
			ReceiverID: receiverID,
			Status:     entity.ReqStatusPending,
		}

		if err = tx.Requests().Create(ctx, &req); err != nil {

			return err
		}

		return insertOutboxEvent(ctx, tx, req.ID.String(), "friendship_request",
			domain.EventFriendRequestSent,
			domain.FriendRequestSentPayload{
				RequestID:  req.ID.String(),
				SenderID:   senderID.String(),
				ReceiverID: receiverID.String(),
			},
		)
	})

	if err != nil {

		return err
	}
	ctxlog.From(ctx).With(
		slog.String("op", "FriendshipUseCase.SendRequest"),
		slog.String("sender_id", senderID.String()),
		slog.String("receiver_id", receiverID.String()),
	).Info("friend request sent")

	return nil
}

func (uc *FriendshipUseCase) AcceptRequest(ctx context.Context, receiverID, senderID uuid.UUID) error {
	err := uc.uow.Do(ctx, func(tx domain.Tx) error {
		if err := tx.Requests().UpdateStatus(ctx, senderID, receiverID, entity.ReqStatusAccepted); err != nil {

			return err
		}

		friendship := entity.Friendship{
			ID:      uuid.New(),
			User1ID: senderID,
			User2ID: receiverID,
		}

		if err := tx.Friendships().Create(ctx, &friendship); err != nil {

			return err
		}

		return insertOutboxEvent(ctx, tx, friendship.ID.String(), "friendship",
			domain.EventFriendshipCreated,
			domain.FriendshipCreatedPayload{
				FriendshipID: friendship.ID.String(),
				User1ID:      friendship.User1ID.String(),
				User2ID:      friendship.User2ID.String(),
			},
		)
	})

	if err != nil {

		return err
	}

	ctxlog.From(ctx).With(
		slog.String("op", "FriendshipUseCase.AcceptRequest"),
		slog.String("sender_id", senderID.String()),
		slog.String("receiver_id", receiverID.String()),
	).Info("friend request accepted")

	return nil
}

func (uc *FriendshipUseCase) RejectRequest(ctx context.Context, receiverID, senderID uuid.UUID) error {
	err := uc.uow.Do(ctx, func(tx domain.Tx) error {
		if err := tx.Requests().UpdateStatus(ctx, senderID, receiverID, entity.ReqStatusRejected); err != nil {

			return err
		}

		return insertOutboxEvent(ctx, tx, senderID.String(), "friendship_status",
			domain.EventFriendRequestRejected,
			domain.FriendRequestRejectedPayload{
				SenderID:   senderID.String(),
				ReceiverID: receiverID.String(),
			},
		)
	})

	if err != nil {

		return err
	}

	ctxlog.From(ctx).With(
		slog.String("op", "FriendshipUseCase.RejectRequest"),
		slog.String("sender_id", senderID.String()),
		slog.String("receiver_id", receiverID.String()),
	).Info("friend request rejected")

	return nil
}

func (uc *FriendshipUseCase) CancelRequest(ctx context.Context, senderID, receiverID uuid.UUID) error {
	err := uc.uow.Do(ctx, func(tx domain.Tx) error {
		if err := tx.Requests().UpdateStatus(ctx, senderID, receiverID, entity.ReqStatusCancelled); err != nil {

			return err
		}

		return insertOutboxEvent(ctx, tx, senderID.String(), "friendship_request",
			domain.EventFriendRequestCancelled,
			domain.FriendRequestCancelledPayload{
				SenderID:   senderID.String(),
				ReceiverID: receiverID.String(),
			},
		)
	})

	if err != nil {

		return err
	}

	ctxlog.From(ctx).With(
		slog.String("op", "FriendshipUseCase.CancelRequest"),
		slog.String("sender_id", senderID.String()),
		slog.String("receiver_id", receiverID.String()),
	).Info("friend request cancelled")

	return nil
}

func (uc *FriendshipUseCase) DeleteFriendship(ctx context.Context, userID, friendID uuid.UUID) error {
	err := uc.uow.Do(ctx, func(tx domain.Tx) error {
		if err := tx.Friendships().Delete(ctx, userID, friendID); err != nil {

			return err
		}

		return insertOutboxEvent(ctx, tx, userID.String(), "friendship",
			domain.EventFriendshipDeleted,
			domain.FriendshipDeletedPayload{
				UserID:   userID.String(),
				FriendID: friendID.String(),
			},
		)
	})

	if err != nil {

		return err
	}

	ctxlog.From(ctx).With(
		slog.String("op", "FriendshipUseCase.DeleteFriendship"),
		slog.String("user_id", userID.String()),
		slog.String("friend_id", friendID.String()),
	).Info("friendship deleted")

	return nil
}

func (uc *FriendshipUseCase) GetFriendsIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return uc.uow.Friendships().GetFriendIDs(ctx, userID)

}

func (uc *FriendshipUseCase) AreFriends(ctx context.Context, user1, user2 uuid.UUID) (bool, error) {
	return uc.uow.Friendships().Exists(ctx, user1, user2)
}

func marshalPayload(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, commonapperr.Internal("marshal outbox payload", err)
	}
	return b, nil
}

func insertOutboxEvent(ctx context.Context, tx domain.Tx, aggregateID, aggregateType, eventType string, payload any) error {
	b, err := marshalPayload(payload)
	if err != nil {
		return err
	}
	return tx.Outbox().Insert(ctx, &outbox.OutboxEvent{
		ID:            uuid.New(),
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		EventType:     eventType,
		Payload:       b,
	})
}
