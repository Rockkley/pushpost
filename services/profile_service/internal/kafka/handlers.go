package kafka

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rockkley/pushpost/services/profile_service/internal/domain/events"
)

type UserCreatedHandler interface {
	Handle(ctx context.Context, event domain.UserCreatedEvent) error
}

type UserCreatedProcessor struct {
	log *slog.Logger
}

func NewUserCreatedProcessor(log *slog.Logger) *UserCreatedProcessor {
	return &UserCreatedProcessor{
		log: log.With("handler", "user_created"),
	}
}

func (h *UserCreatedProcessor) Handle(ctx context.Context, evt domain.UserCreatedEvent) error {
	h.log.Info("handling user.created",
		slog.String("user_id", evt.UserID),
	)
	fmt.Println("HANDLEING WORKS")
	h.log.Info("HANDLEING WORKS")
	//
	// profileRepo.CreateProfile(ctx, evt.UserID)

	return nil
}
