package storage

import (
	"github.com/google/uuid"
	"pushpost/internal/services/user_service/entity"
)

type UserRepository interface {
	RegisterUser(user *entity.User) error
	GetUserByEmail(email string) (*entity.User, error)
	GetUserByUUID(uuid uuid.UUID) (*entity.User, error)
	GetFriends(userUUID uuid.UUID) ([]entity.User, error)
	AddFriend(userUUID uuid.UUID, friendEmail string) error
}
