package storage

import (
	"github.com/google/uuid"
	"pushpost/internal/services/user_service/domain/dto"
	"pushpost/internal/services/user_service/entity"
)

type UserRepository interface {
	RegisterUser(user *entity.User) error
	GetUserByEmail(email string) (*entity.User, error)
	GetUserByUUID(uuid uuid.UUID) (*entity.User, error)
	GetFriends(userUUID uuid.UUID) ([]entity.User, error)
	AddFriend(userUUID uuid.UUID, friendEmail string) error
	DeleteFriend(dto *dto.DeleteFriendDTO) error
}

type FriendRequestRepository interface {
	CreateFriendshipRequest(request entity.FriendshipRequest) error
	GetFriendshipRequestsByRecipientUUID(recipientUUID uuid.UUID) ([]entity.FriendshipRequest, error)
	UpdateFriendshipRequestStatus(dto dto.UpdateFriendshipRequestDto) error
	DeleteFriendshipRequest(requestID uuid.UUID) error
}
