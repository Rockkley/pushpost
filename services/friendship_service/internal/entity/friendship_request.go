package entity

import (
	"github.com/google/uuid"
	"time"
)

type FriendshipReqStatus string

const (
	ReqStatusPending   FriendshipReqStatus = "pending"
	ReqStatusAccepted  FriendshipReqStatus = "accepted"
	ReqStatusRejected  FriendshipReqStatus = "rejected"
	ReqStatusCancelled FriendshipReqStatus = "cancelled"
)

type FriendshipRequest struct {
	ID         uuid.UUID
	SenderID   uuid.UUID
	ReceiverID uuid.UUID
	Status     FriendshipReqStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
