package entity

import (
	"github.com/google/uuid"
	"time"
)

type Friendship struct {
	ID        uuid.UUID
	User1ID   uuid.UUID
	User2ID   uuid.UUID
	CreatedAt time.Time
}
