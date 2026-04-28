package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/notification_service/internal/entity"
)

type NotificationUseCase interface {
	CreateAndDeliver(ctx context.Context, n *entity.Notification) error
	GetForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	GetPreferences(ctx context.Context, userID uuid.UUID) ([]*entity.NotificationPreference, error)
	SetPreference(ctx context.Context, pref *entity.NotificationPreference) error
	GenerateTelegramLinkCode(ctx context.Context, userID uuid.UUID) (string, error)
	BindTelegram(ctx context.Context, code string, chatID int64, username string) error
	UnbindTelegram(ctx context.Context, userID uuid.UUID) error
}

type TelegramLinker interface {
	GenerateTelegramLinkCode(ctx context.Context, userID uuid.UUID) (string, error)
	BindTelegram(ctx context.Context, code string, chatID int64, username string) error
	UnbindTelegram(ctx context.Context, userID uuid.UUID) error
}
