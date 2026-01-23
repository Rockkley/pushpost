package domain

import (
	"errors"
	"time"
)

type RelationshipStatus string

const (
	StatusAccepted RelationshipStatus = "accepted"
	StatusPending  RelationshipStatus = "pending"
)

type Relationship struct {
	ID        int64
	UserID    int64
	FriendID  int64
	Status    RelationshipStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	ErrRelationshipAlreadyExists = errors.New("relationship already exists")
	ErrNotFound                  = errors.New("relationship not found")
	ErrPendingRequestsExists     = errors.New("request already sent")
	ErrCannotBefriendYourself    = errors.New("cannot befriend yourself")
	ErrNotFriends                = errors.New("not friends")
	ErrRequestIsNotPending       = errors.New("request is not pending")
)
