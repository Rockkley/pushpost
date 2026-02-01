package domain

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

type User struct {
	Id           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

func (u *User) DisplayName() string {
	if u.IsDeleted() {
		return "[DELETED USER]"
	}
	return u.Username
}

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrInvalidPassword       = errors.New("password must be at least 8 characters long")
	ErrInvalidUsername       = errors.New("username must be 3-30 characters long")
)
