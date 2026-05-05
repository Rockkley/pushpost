package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/friendship_service/internal/entity"
	"time"
)

type FriendshipRequestRepository interface {
	Create(ctx context.Context, request *entity.FriendshipRequest) error
	FindPending(ctx context.Context, senderID, receiverID uuid.UUID) (*entity.FriendshipRequest, error)
	FindPendingBetween(ctx context.Context, user1, user2 uuid.UUID) (*entity.FriendshipRequest, error)
	UpdateStatus(ctx context.Context, senderID, receiverID uuid.UUID, status entity.FriendshipReqStatus) error
	HasRecentRejected(ctx context.Context, senderID, receiverID uuid.UUID, since time.Time) (bool, error)
	GetIncoming(ctx context.Context, receiverID uuid.UUID) ([]*entity.FriendshipRequest, error)
	GetOutgoing(ctx context.Context, senderID uuid.UUID) ([]*entity.FriendshipRequest, error)
}

type FriendshipRepository interface {
	Create(ctx context.Context, friendship *entity.Friendship) error
	Exists(ctx context.Context, user1, user2 uuid.UUID) (bool, error)
	Delete(ctx context.Context, userID, friendID uuid.UUID) error
	GetFriendIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

type BlockRepository interface {
	Create(ctx context.Context, block *entity.Block) error
	Delete(ctx context.Context, userID, targetID uuid.UUID) error
	Exists(ctx context.Context, userID, targetID uuid.UUID) (bool, error)
	GetBlockedUserIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}
