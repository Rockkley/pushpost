package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/friendship_service/internal/entity"
)

type FriendshipRequestRepository interface {
	Create(ctx context.Context, request *entity.FriendshipRequest) error
	FindPending(ctx context.Context, senderID, receiverID uuid.UUID) (*entity.FriendshipRequest, error)
	FindPendingBetween(ctx context.Context, user1, user2 uuid.UUID) (*entity.FriendshipRequest, error)
	UpdateStatus(ctx context.Context, senderID, receiverID uuid.UUID, status entity.FriendshipReqStatus) error
	Delete(ctx context.Context, senderID, receiverID uuid.UUID) error
}

type FriendshipRepository interface {
	Create(ctx context.Context, friendship *entity.Friendship) error
	Exists(ctx context.Context, user1, user2 uuid.UUID) (bool, error)
	Delete(ctx context.Context, userID, user2 uuid.UUID) error
	GetFriendIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}
