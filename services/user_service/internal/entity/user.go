package entity

import (
	"github.com/google/uuid"
	"time"
)

const (
	StatusInactive = "inactive"
	StatusActive   = "active"
	StatusBlocked  = "blocked"
	StatusDeleted  = "deleted"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"password_hash"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

func (u *User) IsDeleted() bool { return u.DeletedAt != nil }
func (u *User) IsActive() bool  { return u.Status == StatusActive }
