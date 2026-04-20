package dto

import (
	"errors"
	"github.com/google/uuid"
)

var (
	ErrInvalidUserID      = errors.New("invalid user ID")
	ErrInvalidOtherUserID = errors.New("invalid other user ID")
	ErrInvalidLimit       = errors.New("limit must be greater than 0")
	ErrInvalidOffset      = errors.New("offset cannot be negative")
)

type GetConversationDTO struct {
	UserID      uuid.UUID
	OtherUserID uuid.UUID
	Limit       int
	Offset      int
}

func (dto *GetConversationDTO) Validate() error {
	if dto.UserID == uuid.Nil {
		return ErrInvalidUserID
	}

	if dto.OtherUserID == uuid.Nil {
		return ErrInvalidOtherUserID
	}

	if dto.Limit <= 0 {
		return ErrInvalidLimit
	}

	if dto.Offset < 0 {
		return ErrInvalidOffset
	}

	return nil
}
