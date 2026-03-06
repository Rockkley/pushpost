package outbox

import (
	"github.com/google/uuid"
	"time"
)

type OutboxEvent struct {
	ID            uuid.UUID
	AggregateID   string
	AggregateType string
	EventType     string
	Payload       []byte
	Status        string
	CreatedAt     *time.Time
	ProcessedAt   *time.Time
}
