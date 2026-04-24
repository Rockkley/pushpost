package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	streamKeyPrefix = "feed_stream:"
	streamMaxLen    = 1000 // max number of events per user stream (old will be trimmed as new are added)
)

type RedisStreamsNotifier struct {
	rdb *redis.Client
	log *slog.Logger
}

func NewRedisStreamsNotifier(rdb *redis.Client, log *slog.Logger) *RedisStreamsNotifier {
	return &RedisStreamsNotifier{rdb: rdb, log: log.With("component", "streams_notifier")}
}

func (n *RedisStreamsNotifier) Publish(ctx context.Context, userIDs []uuid.UUID, event FeedEvent) error {
	if len(userIDs) == 0 {
		return nil
	}

	b, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal feed event: %w", err)
	}

	pipe := n.rdb.Pipeline()
	for _, userID := range userIDs {
		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: streamKeyPrefix + userID.String(),
			MaxLen: streamMaxLen,
			Approx: true,
			Values: map[string]any{
				"type":    string(event.Type),
				"payload": string(b),
			},
		})
	}

	if _, err = pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis pipeline publish: %w", err)
	}
	return nil
}

func StreamKey(userID uuid.UUID) string {
	return streamKeyPrefix + userID.String()
}
