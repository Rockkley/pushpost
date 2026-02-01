package domain

import "github.com/google/uuid"

type Session struct {
	SessionID uuid.UUID
	UserID    uuid.UUID
	DeviceID  uuid.UUID
	Expires   int64
}
