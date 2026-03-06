package entity

import (
	"errors"
	"github.com/google/uuid"
	"time"
	"unicode/utf8"
)

const (
	MinMessageLength = 1 //fixme parse from config
	MaxMessageLength = 10000
)

type Message struct {
	Id         uuid.UUID
	SenderID   uuid.UUID
	ReceiverID uuid.UUID
	Content    string
	CreatedAt  time.Time
	ReadAt     *time.Time
}

func (m *Message) IsRead() bool {
	return m.ReadAt != nil
}

var (
	ErrMessageNotFound    = errors.New("message_service not found")
	ErrCannotMessageSelf  = errors.New("cannot send message_service to yourself")
	ErrMessageTooLong     = errors.New("message_service exceeds maximum length of 10000 characters") //fixme parse from var
	ErrMessageEmpty       = errors.New("message_service cannot be empty")
	ErrCannotMessageUser  = errors.New("you cannot send messages to this user_service")
	ErrReceiverNotFound   = errors.New("receiver not found")
	ErrReceiverDeleted    = errors.New("cannot message_service deleted user_service")
	ErrSenderDeleted      = errors.New("your account has been deleted")
	ErrNotMessageReceiver = errors.New("you are not the receiver of this message_service")
)

func ValidateContent(content string) error {

	length := utf8.RuneCountInString(content)
	if length > MaxMessageLength {
		return ErrMessageTooLong
	}

	if length < MinMessageLength {
		return ErrMessageEmpty
	}

	return nil
}
