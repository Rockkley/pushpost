package friendship

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/services/friendship/transport/http/dto"
)

type FriendService interface {
	SendRequest(ctx context.Context, requestDTO dto.SendRequestDTO) error
	AcceptRequest(ctx context.Context, dto dto.AcceptRequestDTO) error
	RejectRequest(ctx context.Context, dto dto.RejectRequestDTO) error
	CancelRequest(ctx context.Context, dto dto.CancelRequestDTO) error
	DeleteFriendship(ctx context.Context, dto dto.DeleteFriendshipDTO) error
	GetFriendsIDs(ctx context.Context, userID uuid.UUID) ([]*uuid.UUID, error)
	AreFriends(ctx context.Context, user1, user2 uuid.UUID) (bool, error)
}
