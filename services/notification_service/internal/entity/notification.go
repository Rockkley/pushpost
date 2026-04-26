package entity

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string
type Channel string

const (
	TypeFriendRequestReceived NotificationType = "friend_request.received"
	TypeFriendRequestAccepted NotificationType = "friend_request.accepted"
	TypeFriendRequestRejected NotificationType = "friend_request.rejected"
	TypeMessageReceived       NotificationType = "message.received"
)

const (
	ChannelInApp    Channel = "in_app"
	ChannelTelegram Channel = "telegram"
)

var AllChannels = []Channel{ChannelInApp, ChannelTelegram}

type Notification struct {
	ID        uuid.UUID         `json:"id"`
	UserID    uuid.UUID         `json:"user_id"`
	Type      NotificationType  `json:"type"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	Data      map[string]string `json:"data,omitempty"`
	ReadAt    *time.Time        `json:"read_at,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

func (n *Notification) IsRead() bool { return n.ReadAt != nil }

type NotificationPreference struct {
	UserID  uuid.UUID        `json:"user_id"`
	Type    NotificationType `json:"type"`
	Channel Channel          `json:"channel"`
	Enabled bool             `json:"enabled"`
}

type TelegramBinding struct {
	UserID    uuid.UUID `json:"user_id"`
	ChatID    int64     `json:"chat_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}
