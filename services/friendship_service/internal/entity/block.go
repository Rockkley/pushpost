package entity

import (
	"github.com/google/uuid"
	"time"
)

type Block struct {
	UserID    uuid.UUID
	TargetID  uuid.UUID
	CreatedAt time.Time
}
