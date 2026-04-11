package kafka

import (
	"context"
	"fmt"
	domain2 "github.com/rockkley/pushpost/services/profile_service/internal/domain/events"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/profile_service/internal/domain"
	"github.com/rockkley/pushpost/services/profile_service/internal/entity"
)

type UserCreatedHandler interface {
	Handle(ctx context.Context, event domain2.UserCreatedEvent) error
}

type UserCreatedProcessor struct {
	uc  domain.ProfileUseCaseInterface
	log *slog.Logger
}

func NewUserCreatedProcessor(uc domain.ProfileUseCaseInterface, log *slog.Logger) *UserCreatedProcessor {
	return &UserCreatedProcessor{
		uc:  uc,
		log: log.With("handler", "user_created"),
	}
}

func (h *UserCreatedProcessor) Handle(ctx context.Context, evt domain2.UserCreatedEvent) error {
	userID, err := uuid.Parse(evt.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id %q: %w", evt.UserID, err)
	}

	profile := &entity.Profile{
		UserID:    userID,
		Username:  evt.Username,
		CreatedAt: time.Now(),
	}

	if err = h.uc.CreateProfile(ctx, profile); err != nil {
		h.log.Error("failed to create profile",
			slog.String("user_id", evt.UserID),
			slog.Any("error", err),
		)
		return err
	}

	h.log.Info("profile created",
		slog.String("user_id", evt.UserID),
		slog.String("username", evt.Username),
	)
	return nil
}
