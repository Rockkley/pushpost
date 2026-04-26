package delivery

import (
	"context"

	"github.com/rockkley/pushpost/services/notification_service/internal/entity"
)

type Deliverer interface {
	Channel() entity.Channel
	Deliver(ctx context.Context, n *entity.Notification) error
}
