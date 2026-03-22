package domain

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	"github.com/rockkley/pushpost/services/friendship_service/internal/entity"
	"github.com/rockkley/pushpost/services/friendship_service/internal/repository"
)

type FriendshipUseCase interface {
	SendRequest(ctx context.Context, senderID, receiverID uuid.UUID) error
	AcceptRequest(ctx context.Context, receiverID, senderID uuid.UUID) error
	RejectRequest(ctx context.Context, receiverID, senderID uuid.UUID) error
	CancelRequest(ctx context.Context, senderID, receiverID uuid.UUID) error
	DeleteFriendship(ctx context.Context, userID, friendID uuid.UUID) error
	GetFriendIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	AreFriends(ctx context.Context, user1, user2 uuid.UUID) (bool, error)
	GetFriendshipStatus(ctx context.Context, viewerID, targetID uuid.UUID) (*entity.FriendshipStatus, error)
	GetIncomingRequests(ctx context.Context, userID uuid.UUID) ([]*entity.FriendshipRequest, error)
}

type Tx interface {
	Requests() repository.FriendshipRequestRepository
	Friendships() repository.FriendshipRepository
	Outbox() outbox.WriterInterface
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(tx Tx) error) error
	Requests() repository.FriendshipRequestRepository
	Friendships() repository.FriendshipRepository
}
