package realtime

import (
	"context"

	"github.com/google/uuid"
)

type EventType string

const (
	EventPostAdded     EventType = "post_added"
	EventPostUpdated   EventType = "post_updated"
	EventPostDeleted   EventType = "post_deleted"
	EventFriendRemoved EventType = "friend_removed"
	EventBulkNewPosts  EventType = "bulk_new_posts"
)

type FeedEvent struct {
	Type     EventType `json:"type"`
	PostID   string    `json:"post_id,omitempty"`
	FriendID string    `json:"friend_id,omitempty"`
	Version  int       `json:"version,omitempty"`
	PostIDs  []string  `json:"post_ids,omitempty"`
}

type Notifier interface {
	Publish(ctx context.Context, userIDs []uuid.UUID, event FeedEvent) error
}
