package postgres

import (
	"context"
	"github.com/google/uuid"
)

type FriendshipRepository struct {
}

func (r FriendshipRepository) DeleteFriendship(ctx context.Context, userID, friendID uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (r FriendshipRepository) AreFriends(ctx context.Context, user1, user2 uuid.UUID) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (r FriendshipRepository) GetFriendsIds(ctx context.Context, userID uuid.UUID) ([]*uuid.UUID, error) {
	//TODO implement me
	panic("implement me")
}

func (r FriendshipRepository) RequestExists(ctx context.Context, senderID, receiverID uuid.UUID) (bool, error) {
	//TODO implement me
	panic("implement me")
}
