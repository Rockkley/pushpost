package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rockkley/pushpost/services/post_service/internal/entity"
)

const feedKeyPrefix = "feed:"

type FeedCache struct {
	rdb *redis.Client
}

func NewFeedCache(rdb *redis.Client) *FeedCache {
	return &FeedCache{rdb: rdb}
}

func (c *FeedCache) GetFeed(ctx context.Context, userID uuid.UUID) ([]*entity.Post, error) {
	data, err := c.rdb.Get(ctx, feedKey(userID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {

			return nil, fmt.Errorf("cache miss")
		}

		return nil, err
	}

	var posts []*entity.Post

	if err = json.Unmarshal(data, &posts); err != nil {

		return nil, err
	}

	return posts, nil
}

func (c *FeedCache) SetFeed(ctx context.Context, userID uuid.UUID, posts []*entity.Post, ttl time.Duration) error {
	data, err := json.Marshal(posts)

	if err != nil {

		return err
	}

	return c.rdb.Set(ctx, feedKey(userID), data, ttl).Err()
}

func (c *FeedCache) InvalidateFeed(ctx context.Context, userIDs []uuid.UUID) error {
	keys := make([]string, len(userIDs))
	for i, id := range userIDs {
		keys[i] = feedKey(id)
	}

	return c.rdb.Del(ctx, keys...).Err()
}

func feedKey(userID uuid.UUID) string {
	return feedKeyPrefix + userID.String()
}
