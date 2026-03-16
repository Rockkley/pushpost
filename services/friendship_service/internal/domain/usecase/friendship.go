package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/friendship_service/internal/repository"
	"github.com/rockkley/pushpost/services/friendship_service/internal/transport/http/dto"
)

type FriendshipUseCase struct {
	friendshipRepo repository.FriendshipRepository
}

func (f FriendshipUseCase) SendRequest(ctx context.Context, requestDTO dto.SendRequestDTO) error {
	//TODO implement me
	panic("implement me")
}

func (f FriendshipUseCase) AcceptRequest(ctx context.Context, dto dto.AcceptRequestDTO) error {
	//TODO implement me
	panic("implement me")
}

func (f FriendshipUseCase) RejectRequest(ctx context.Context, dto dto.RejectRequestDTO) error {
	//TODO implement me
	panic("implement me")
}

func (f FriendshipUseCase) CancelRequest(ctx context.Context, dto dto.CancelRequestDTO) error {
	//TODO implement me
	panic("implement me")
}

func (f FriendshipUseCase) DeleteFriendship(ctx context.Context, dto dto.DeleteFriendshipDTO) error {
	//TODO implement me
	panic("implement me")
}

func (f FriendshipUseCase) GetFriendsIDs(ctx context.Context, userID uuid.UUID) ([]*uuid.UUID, error) {
	//TODO implement me
	panic("implement me")
}

func (f FriendshipUseCase) AreFriends(ctx context.Context, user1, user2 uuid.UUID) (bool, error) {
	//TODO implement me
	panic("implement me")
}
