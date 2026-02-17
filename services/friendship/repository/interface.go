package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/friendship/entity"
)

type FriendshipRequestRepository interface {
	CreateRequest(ctx context.Context, request *entity.FriendshipRequest) error
	AcceptRequest(ctx context.Context, senderID, receiverID uuid.UUID) error
	RejectRequest(ctx context.Context, senderID, receiverID uuid.UUID) error
	CancelRequest(ctx context.Context, senderID, receiverID uuid.UUID) error
	RequestExists(ctx context.Context, senderID, receiverID uuid.UUID) (bool, error)
}
type FriendshipRepository interface {
	DeleteFriendship(ctx context.Context, userID, friendID uuid.UUID) error
	AreFriends(ctx context.Context, user1, user2 uuid.UUID) (bool, error)
	GetFriendsIds(ctx context.Context, userID uuid.UUID) ([]*uuid.UUID, error)
}
