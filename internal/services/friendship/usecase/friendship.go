package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/apperror"
	"github.com/rockkley/pushpost/internal/services/friendship/entity"
	"github.com/rockkley/pushpost/internal/services/friendship/repository"
	"github.com/rockkley/pushpost/internal/services/friendship/transport/http/dto"
	"time"
)

type FriendService struct {
	userRepo       repository.UserRepository
	friendshipRepo repository.FriendshipRepository
	requestsRepo   repository.FriendshipRequestRepository
}

func (f *FriendService) SendRequest(ctx context.Context, requestDTO dto.SendRequestDTO) error {
	if err := f.ValidateSendRequest(ctx, requestDTO); err != nil {
		return err
	}

	request := entity.FriendshipRequest{
		ID:         uuid.New(),
		SenderID:   requestDTO.SenderID,
		ReceiverID: requestDTO.ReceiverID,
		Status:     entity.StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := f.requestsRepo.CreateRequest(ctx, &request); err != nil {
		return err
	}

	return nil
}

func (f *FriendService) AcceptRequest(ctx context.Context, d dto.AcceptRequestDTO) error {
	//TODO implement me
	panic("implement me")
}

func (f *FriendService) RejectRequest(ctx context.Context, d dto.RejectRequestDTO) error {
	//TODO implement me
	panic("implement me")
}

func (f *FriendService) CancelRequest(ctx context.Context, d dto.CancelRequestDTO) error {
	//TODO implement me
	panic("implement me")
}

func (f *FriendService) DeleteFriendship(ctx context.Context, d dto.DeleteFriendshipDTO) error {
	//TODO implement me
	panic("implement me")
}

func (f *FriendService) GetFriendsIDs(ctx context.Context, userID uuid.UUID) ([]*uuid.UUID, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FriendService) AreFriends(ctx context.Context, user1, user2 uuid.UUID) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FriendService) ValidateSendRequest(ctx context.Context, requestDTO dto.SendRequestDTO) error {
	if requestDTO.SenderID == requestDTO.ReceiverID {
		return apperror.BadRequest(apperror.CodeCannotBefriendSelf, "cannot send friend request to yourself")
	}

	receiver, err := f.userRepo.FindByID(ctx, requestDTO.ReceiverID)
	if err != nil {
		return err
	}

	if receiver == nil {
		return apperror.NotFound(apperror.CodeUserNotFound, "receiver user_service not found")
	}

	if receiver.IsDeleted() {
		return apperror.BadRequest(apperror.CodeUserDeleted, "cannot send friend request to deleted user_service")
	}

	areFriends, err := f.AreFriends(ctx, requestDTO.SenderID, requestDTO.ReceiverID)
	if err != nil {
		return err
	}

	if areFriends {
		return apperror.BadRequest(apperror.CodeAlreadyFriends, "users are already friends")
	}

	existsForward, err := f.requestsRepo.RequestExists(ctx, requestDTO.SenderID, requestDTO.ReceiverID)
	if err != nil {
		return err
	}

	if existsForward {
		return apperror.BadRequest(apperror.CodeFriendRequestExists, "friend request already exists between users")
	}

	existsBackward, err := f.requestsRepo.RequestExists(ctx, requestDTO.ReceiverID, requestDTO.SenderID)
	if err != nil {
		return err
	}

	if existsBackward {
		return apperror.BadRequest(apperror.CodeFriendRequestExists, "friend request already exists between users")
	}

	return nil
}
