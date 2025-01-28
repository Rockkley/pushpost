package repository

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"pushpost/internal/services/user_service/domain/dto"
	"pushpost/internal/services/user_service/entity"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(DB *gorm.DB) *UserRepository {
	return &UserRepository{DB: DB}
}

func (r *UserRepository) RegisterUser(user *entity.User) error {

	return r.DB.Create(&user).Error
}

func (r *UserRepository) GetUserByEmail(email string) (*entity.User, error) {
	var user entity.User
	fmt.Println("EMAIL", email)

	if err := r.DB.Where("email = ?", email).First(&user).Error; err != nil {

		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserByUUID(uuid uuid.UUID) (*entity.User, error) {
	var user entity.User

	if err := r.DB.Where("uuid = ?", uuid).First(&user).Error; err != nil {

		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetFriends(userUUID uuid.UUID) ([]entity.User, error) {
	var friends []entity.User

	// Get the user's ID first
	user, err := r.GetUserByUUID(userUUID)
	if err != nil {
		return nil, err
	}

	// Query to get all friends through the friendships table
	err = r.DB.
		Joins("JOIN friendships ON users.id = friendships.friend_id").
		Where("friendships.user_id = ? AND friendships.deleted_at IS NULL", user.ID).
		Or("friendships.friend_id = ? AND friendships.deleted_at IS NULL", user.ID).
		Where("users.deleted_at IS NULL").
		Find(&friends).Error

	return friends, err
}
func (r *UserRepository) AddFriend(userUUID uuid.UUID, friendEmail string) error {
	var existingFriendship entity.Friendship

	user, err := r.GetUserByUUID(userUUID)

	if err != nil {
		fmt.Println("user not found by UUID")
		return err
	}
	friend, err := r.GetUserByEmail(friendEmail)

	if err != nil {

		return err
	}
	//Check if friendship already exists in either direction
	result := r.DB.Where(
		"(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		user.ID, friend.ID, friend.ID, user.ID,
	).First(&existingFriendship)

	if result.Error == nil {

		return fmt.Errorf("friendship already exists")
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {

		return result.Error
	}

	friendship := entity.Friendship{UserID: user.ID, FriendID: friend.ID}

	return r.DB.Create(&friendship).Error
}

func (r *UserRepository) DeleteFriend(dto *dto.DeleteFriendDTO) error {

	user, err := r.GetUserByUUID(dto.UserUUID)

	if err != nil {
		fmt.Println("user not found by UUID")
		return err
	}
	friend, err := r.GetUserByEmail(dto.FriendEmail)

	if err != nil {

		return err
	}
	fmt.Println("USER ID ", user.ID, " FRIEND ID ", friend.ID)
	//Check if friendship exists in either direction
	result := r.DB.Where(
		"(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		user.ID, friend.ID, friend.ID, user.ID,
	).First(&entity.Friendship{})

	if result.Error != nil {
		fmt.Println(result.Error)
		return fmt.Errorf("friendship not exists")
	}

	//if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
	//
	//	return result.Error
	//}

	//friendship := entity.Friendship{UserID: user.ID, FriendID: friend.ID}
	return r.DB.Unscoped().Where(
		"(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		user.ID, friend.ID, friend.ID, user.ID).Delete(&entity.Friendship{}).Error
}
