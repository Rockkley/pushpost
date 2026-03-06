package outbox

import (
	"github.com/google/uuid"
	"time"
)

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusProcessed  = "processed"
)

type OutboxEvent struct {
	ID            uuid.UUID
	AggregateID   string
	AggregateType string
	Attempts      int
	EventType     string
	Payload       []byte
	Status        string
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}
