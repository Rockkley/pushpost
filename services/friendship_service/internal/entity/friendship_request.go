package entity

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

type FriendshipStatus string

const (
	StatusAccepted FriendshipStatus = "accepted"
	StatusPending  FriendshipStatus = "pending"
)

type FriendshipRequest struct {
	ID         uuid.UUID
	SenderID   uuid.UUID
	ReceiverID uuid.UUID
	Status     FriendshipStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

var (
	ErrRelationshipAlreadyExists = errors.New("relationship already exists")
	ErrNotFound                  = errors.New("relationship not found")
	ErrPendingRequestsExists     = errors.New("request already sent")
	ErrCannotBefriendYourself    = errors.New("cannot befriend yourself")
	ErrNotFriends                = errors.New("not friends")
	ErrRequestIsNotPending       = errors.New("request is not pending")
)
