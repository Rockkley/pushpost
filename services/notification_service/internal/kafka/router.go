package kafka

import (
	"context"
	"log/slog"
)

type EventRouter struct {
	handlers *Handlers
	log      *slog.Logger
}

func NewRouter(handlers *Handlers, log *slog.Logger) *EventRouter {
	return &EventRouter{handlers: handlers, log: log.With("component", "notification_event_router")}
}

func (r *EventRouter) Route(ctx context.Context, topic string, env Envelope) error {
	eventType := env.EventType
	if eventType == "" {
		// Compatibility fallback: some producers rely on Kafka topic as event type.
		eventType = topic
	}

	if len(env.Payload) == 0 {
		r.log.Warn("empty event payload, skipping", slog.String("event_type", eventType), slog.String("topic", topic))
		return nil
	}

	switch eventType {
	case "friendship_request.sent":
		return r.handlers.HandleFriendRequestSent(ctx, env.Payload)
	case "friendship.created":
		return r.handlers.HandleFriendshipCreated(ctx, env.Payload)
	case "friendship_request.rejected":
		return r.handlers.HandleFriendRequestRejected(ctx, env.Payload)
	case "message.sent":
		return r.handlers.HandleMessageSent(ctx, env.Payload)
	default:
		r.log.Debug("unhandled event type, skipping", slog.String("event_type", eventType), slog.String("topic", topic))
		return nil
	}
}
