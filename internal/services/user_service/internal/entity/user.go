package entity

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

var (
	ErrUserNotFound = errors.New("user_service not found")
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
