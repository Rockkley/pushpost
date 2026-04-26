package usecase

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	"github.com/rockkley/pushpost/services/notification_service/internal/delivery"
	"github.com/rockkley/pushpost/services/notification_service/internal/domain"
	"github.com/rockkley/pushpost/services/notification_service/internal/entity"
	"github.com/rockkley/pushpost/services/notification_service/internal/repository"
)

var _ domain.NotificationUseCase = (*NotificationUseCase)(nil)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type NotificationUseCase struct {
	repo       repository.NotificationRepository
	prefRepo   repository.PreferenceRepository
	binder     *TelegramBinder
	deliverers []delivery.Deliverer
	log        *slog.Logger
}

func NewNotificationUseCase(repo repository.NotificationRepository, prefRepo repository.PreferenceRepository, binder *TelegramBinder, deliverers []delivery.Deliverer, log *slog.Logger) *NotificationUseCase {
	return &NotificationUseCase{repo: repo, prefRepo: prefRepo, binder: binder, deliverers: deliverers, log: log.With("component", "notification_usecase")}
}

func (uc *NotificationUseCase) CreateAndDeliver(ctx context.Context, n *entity.Notification) error {
	log := ctxlog.From(ctx).With(slog.String("op", "NotificationUseCase.CreateAndDeliver"), slog.String("user_id", n.UserID.String()), slog.String("type", string(n.Type)))
	if err := uc.repo.Create(ctx, n); err != nil {
		log.Error("failed to persist notification", slog.Any("error", err))
		return err
	}
	enabledChannels, err := uc.prefRepo.GetEnabled(ctx, n.UserID, n.Type)
	if err != nil {
		log.Warn("preference lookup failed, delivering to all channels", slog.Any("error", err))
		enabledChannels = entity.AllChannels
	}
	channelSet := make(map[entity.Channel]struct{}, len(enabledChannels))
	for _, ch := range enabledChannels {
		channelSet[ch] = struct{}{}
	}
	for _, d := range uc.deliverers {
		if _, ok := channelSet[d.Channel()]; !ok {
			continue
		}
		if deliverErr := d.Deliver(ctx, n); deliverErr != nil {
			log.Warn("delivery failed", slog.String("channel", string(d.Channel())), slog.Any("error", deliverErr))
		}
	}
	return nil
}

func (uc *NotificationUseCase) GetForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, error) {
	if limit <= 0 || limit > maxLimit {
		limit = defaultLimit
	}
	if offset < 0 {
		offset = 0
	}
	return uc.repo.GetForUser(ctx, userID, limit, offset)
}

func (uc *NotificationUseCase) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return uc.repo.GetUnreadCount(ctx, userID)
}
func (uc *NotificationUseCase) MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	return uc.repo.MarkAsRead(ctx, notificationID, userID)
}
func (uc *NotificationUseCase) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return uc.repo.MarkAllAsRead(ctx, userID)
}
func (uc *NotificationUseCase) GetPreferences(ctx context.Context, userID uuid.UUID) ([]*entity.NotificationPreference, error) {
	return uc.prefRepo.GetAll(ctx, userID)
}
func (uc *NotificationUseCase) SetPreference(ctx context.Context, pref *entity.NotificationPreference) error {
	return uc.prefRepo.Upsert(ctx, pref)
}
func (uc *NotificationUseCase) GenerateTelegramLinkCode(ctx context.Context, userID uuid.UUID) (string, error) {
	return uc.binder.GenerateTelegramLinkCode(ctx, userID)
}
func (uc *NotificationUseCase) BindTelegram(ctx context.Context, code string, chatID int64, username string) error {
	return uc.binder.BindTelegram(ctx, code, chatID, username)
}
func (uc *NotificationUseCase) UnbindTelegram(ctx context.Context, userID uuid.UUID) error {
	return uc.binder.UnbindTelegram(ctx, userID)
}
