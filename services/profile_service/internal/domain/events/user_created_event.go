package domain

const (
	EventUserCreated = "user.created"
)

type Envelope struct {
	EventType string `json:"event_type"`
	Payload   []byte `json:"payload"`
}

type UserCreatedEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}
