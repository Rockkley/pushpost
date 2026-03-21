package entity

import "github.com/google/uuid"

type Profile struct {
	UserID       uuid.UUID
	FirstName    string
	SecondName   string
	Birthdate    string
	AvatarLink   string
	Bio          string
	TelegramLink string
}
