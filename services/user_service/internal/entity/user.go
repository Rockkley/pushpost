package entity

import (
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
