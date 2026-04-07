package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/rockkley/pushpost/services/profile_service/internal/domain/events"
)

type Router struct {
	userCreatedHandler UserCreatedHandler
	log                *slog.Logger
}

func NewRouter(userCreatedHandler UserCreatedHandler, log *slog.Logger) *Router {
	return &Router{
		userCreatedHandler: userCreatedHandler,
		log:                log.With("component", "event_router"),
	}
}

func (r *Router) Route(ctx context.Context, env domain.Envelope) error {
	fmt.Println(env)
	switch env.EventType {
	case domain.EventUserCreated:
		fmt.Println("ROUTED")
		var evt domain.UserCreatedEvent
		if err := json.Unmarshal(env.Payload, &evt); err != nil {
			return fmt.Errorf("decode user.created: %w", err)
		}
		return r.userCreatedHandler.Handle(ctx, evt)

	default:
		r.log.Warn("unhandled event", slog.String("event_type", env.EventType))
		return nil
	}
}
