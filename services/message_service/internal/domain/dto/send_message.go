package dto

import (
	"errors"
	"github.com/google/uuid"
)

var (
	ErrInvalidSenderID   = errors.New("invalid sender ID")
	ErrInvalidReceiverID = errors.New("invalid receiver ID")
)

type SendMessageDTO struct {
	SenderID   uuid.UUID
	ReceiverID uuid.UUID
	Content    string
}

func (dto *SendMessageDTO) Validate() error {
	if dto.SenderID == uuid.Nil {

		return ErrInvalidSenderID
	}

	if dto.ReceiverID == uuid.Nil {

		return ErrInvalidReceiverID
	}

	if err := dto.ValidateContent(dto.Content); err != nil {

		return err
	}

	return nil

}

func (dto *SendMessageDTO) ValidateContent(content string) error {
	if len(content) == 0 {

		return errors.New("message content cannot be empty")
	}

	if len(content) > 10000 { //fixme magic number

		return errors.New("message content exceeds maximum length of 10000 characters")
	}

	return nil
}
