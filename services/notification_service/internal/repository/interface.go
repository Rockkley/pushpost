package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/notification_service/internal/entity"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *entity.Notification) error
	GetForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
}

type PreferenceRepository interface {
	GetAll(ctx context.Context, userID uuid.UUID) ([]*entity.NotificationPreference, error)
	GetEnabled(ctx context.Context, userID uuid.UUID, nType entity.NotificationType) ([]entity.Channel, error)
	Upsert(ctx context.Context, pref *entity.NotificationPreference) error
}

type TelegramRepository interface {
	Upsert(ctx context.Context, binding *entity.TelegramBinding) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.TelegramBinding, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}

type LinkCodeStore interface {
	Save(ctx context.Context, code string, userID uuid.UUID, ttl time.Duration) error
	Pop(ctx context.Context, code string) (uuid.UUID, error)
}
