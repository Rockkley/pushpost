package domain

import "encoding/json"

const (
	EventUserCreated = "user.created"
)

type Envelope struct {
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

type UserCreatedEvent struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}
