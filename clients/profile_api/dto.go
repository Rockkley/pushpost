package profile_api

import (
	"github.com/google/uuid"
	"time"
)

type ProfileResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}
