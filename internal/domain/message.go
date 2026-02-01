package domain

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
	ErrMessageNotFound    = errors.New("message not found")
	ErrCannotMessageSelf  = errors.New("cannot send message to yourself")
	ErrMessageTooLong     = errors.New("message exceeds maximum length of 10000 characters") //fixme parse from var
	ErrMessageEmpty       = errors.New("message cannot be empty")
	ErrCannotMessageUser  = errors.New("you cannot send messages to this user")
	ErrReceiverNotFound   = errors.New("receiver not found")
	ErrReceiverDeleted    = errors.New("cannot message deleted user")
	ErrSenderDeleted      = errors.New("your account has been deleted")
	ErrNotMessageReceiver = errors.New("you are not the receiver of this message")
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
