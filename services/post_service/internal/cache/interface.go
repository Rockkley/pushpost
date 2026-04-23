package cache

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/post_service/internal/entity"
	"time"
)

type FeedCache interface {
	GetFeed(ctx context.Context, userID uuid.UUID) ([]*entity.Post, error)
	SetFeed(ctx context.Context, userID uuid.UUID, posts []*entity.Post, ttl time.Duration) error
	InvalidateFeed(ctx context.Context, friendIDs []uuid.UUID) error
}
