package postgres

import (
	"context"
	"github.com/google/uuid"
)

type RelationshipRepository struct {
}

func (r RelationshipRepository) CreateRequest(ctx context.Context, senderID, receiverID uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (r RelationshipRepository) AcceptRequest(ctx context.Context, senderID, receiverID uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (r RelationshipRepository) RejectRequest(ctx context.Context, senderID, receiverID uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (r RelationshipRepository) CancelRequest(ctx context.Context, senderID, receiverID uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (r RelationshipRepository) DeleteFriendship(ctx context.Context, userID, friendID uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (r RelationshipRepository) AreFriends(ctx context.Context, user1, user2 uuid.UUID) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (r RelationshipRepository) GetFriendsIds(ctx context.Context, userID uuid.UUID) ([]*uuid.UUID, error) {
	//TODO implement me
	panic("implement me")
}

func (r RelationshipRepository) RequestExists(ctx context.Context, senderID, receiverID uuid.UUID) (bool, error) {
	//TODO implement me
	panic("implement me")
}
