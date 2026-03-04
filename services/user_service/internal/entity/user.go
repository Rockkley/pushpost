package entity

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	Id           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	DeletedAt    *time.Time `json:"deletedAt,omitempty"`
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}
