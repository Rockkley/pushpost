package inapp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/rockkley/pushpost/services/notification_service/internal/entity"
)

const (
	streamKeyPrefix = "notif_stream:"
	streamMaxLen    = 500
)

// Deliverer publishes a notification to the user's personal Redis Stream.
// The SSE handler reads from this stream and pushes events to connected clients.
type Deliverer struct {
	rdb *goredis.Client
}

func NewDeliverer(rdb *goredis.Client) *Deliverer {
	return &Deliverer{rdb: rdb}
}

func (d *Deliverer) Channel() entity.Channel { return entity.ChannelInApp }

func (d *Deliverer) Deliver(ctx context.Context, n *entity.Notification) error {
	payload, err := json.Marshal(map[string]any{
		"id":         n.ID.String(),
		"type":       string(n.Type),
		"title":      n.Title,
		"body":       n.Body,
		"data":       n.Data,
		"created_at": n.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("marshal in-app notification payload: %w", err)
	}

	return d.rdb.XAdd(ctx, &goredis.XAddArgs{
		Stream: StreamKey(n.UserID),
		MaxLen: streamMaxLen,
		Approx: true,
		Values: map[string]any{
			"type":    string(n.Type),
			"payload": string(payload),
		},
	}).Err()
}

func StreamKey(userID uuid.UUID) string {
	return streamKeyPrefix + userID.String()
}
