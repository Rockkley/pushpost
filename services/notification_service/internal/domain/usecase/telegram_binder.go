package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	apperr "github.com/rockkley/pushpost/services/notification_service/internal/apperror"
	"github.com/rockkley/pushpost/services/notification_service/internal/entity"
	"github.com/rockkley/pushpost/services/notification_service/internal/repository"
)

const linkCodeTTL = 10 * time.Minute

type TelegramBinder struct {
	linkStore    repository.LinkCodeStore
	telegramRepo repository.TelegramRepository
	log          *slog.Logger
}

func NewTelegramBinder(linkStore repository.LinkCodeStore, telegramRepo repository.TelegramRepository, log *slog.Logger) *TelegramBinder {
	return &TelegramBinder{linkStore: linkStore, telegramRepo: telegramRepo, log: log.With("component", "telegram_binder")}
}

func (b *TelegramBinder) GenerateTelegramLinkCode(ctx context.Context, userID uuid.UUID) (string, error) {
	code, err := generateSecureCode()
	if err != nil {
		return "", commonapperr.Internal("generate telegram link code", err)
	}
	if err = b.linkStore.Save(ctx, code, userID, linkCodeTTL); err != nil {
		return "", commonapperr.Internal("save telegram link code", err)
	}
	return code, nil
}

func (b *TelegramBinder) BindTelegram(ctx context.Context, code string, chatID int64, username string) error {
	log := ctxlog.From(ctx).With(slog.String("op", "TelegramBinder.BindTelegram"))
	if code == "" {
		return apperr.InvalidTelegramCode()
	}
	userID, err := b.linkStore.Pop(ctx, code)
	if err != nil {
		log.Warn("invalid or expired link code", slog.String("code", code))
		return apperr.InvalidTelegramCode()
	}
	if err = b.telegramRepo.Upsert(ctx, &entity.TelegramBinding{UserID: userID, ChatID: chatID, Username: username}); err != nil {
		return err
	}
	return nil
}

func (b *TelegramBinder) UnbindTelegram(ctx context.Context, userID uuid.UUID) error {
	return b.telegramRepo.Delete(ctx, userID)
}

func generateSecureCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}
