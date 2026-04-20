package entity

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	MinMessageLength = 1
	MaxMessageLength = 10000
)

type Message struct {
	ID         uuid.UUID  `json:"id"`
	SenderID   uuid.UUID  `json:"sender_id"`
	ReceiverID uuid.UUID  `json:"receiver_id"`
	Content    string     `json:"content"`
	CreatedAt  time.Time  `json:"created_at"`
	ReadAt     *time.Time `json:"read_at,omitempty"`
}

func (m *Message) IsRead() bool { return m.ReadAt != nil }

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

	if length < MinMessageLength {

		return ErrMessageEmpty
	}

	if length > MaxMessageLength {

		return ErrMessageTooLong
	}

	return nil
}
