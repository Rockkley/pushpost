package entity

import (
	"github.com/google/uuid"
	"time"
)

type Profile struct {
	UserID         uuid.UUID
	Username       string
	DisplayName    *string
	FirstName      *string
	LastName       *string
	City           *string
	Country        *string
	BirthDate      *time.Time
	AvatarURL      *string
	AvatarThumbURL *string
	Bio            *string
	TelegramLink   *string
	IsPrivate      bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

func (p *Profile) IsDeleted() bool {
	return p.DeletedAt != nil
}
