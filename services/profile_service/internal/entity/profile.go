package entity

import (
	"github.com/google/uuid"
	"time"
)

type Profile struct {
	UserID       uuid.UUID
	Username     string
	FirstName    string
	SecondName   string
	Birthdate    string
	AvatarLink   string
	Bio          string
	TelegramLink string
	CreatedAt    time.Time
}
