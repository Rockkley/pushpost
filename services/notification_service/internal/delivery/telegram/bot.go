package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rockkley/pushpost/services/notification_service/internal/domain"
)

type Bot struct {
	api    *tgbotapi.BotAPI
	linker domain.TelegramLinker
	log    *slog.Logger
}

var _ MessageSender = (*Bot)(nil)

func NewBot(token string, linker domain.TelegramLinker, log *slog.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	return &Bot{api: api, linker: linker, log: log.With("component", "telegram_bot")}, nil
}

func (b *Bot) Run(ctx context.Context) error {
	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = 30
	updates := b.api.GetUpdatesChan(cfg)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			b.handleUpdate(ctx, update)
		}
	}
}

func (b *Bot) SendMessage(ctx context.Context, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2

	if _, err := b.api.Send(msg); err != nil {
		return fmt.Errorf("send telegram message to chat %d: %w", chatID, err)
	}

	return nil
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	msg := update.Message
	parts := strings.Fields(strings.TrimSpace(msg.Text))

	if len(parts) == 0 || parts[0] != "/start" {
		b.sendText(msg.Chat.ID, "Используйте команду /start <код> для привязки аккаунта PushPost\\.")

		return
	}

	if len(parts) < 2 {
		b.sendText(msg.Chat.ID, "Пожалуйста, укажите код: `/start <код>`")

		return
	}

	username := ""

	if msg.From != nil {
		username = msg.From.UserName
	}

	if err := b.linker.BindTelegram(ctx, parts[1], msg.Chat.ID, username); err != nil {
		b.log.Warn("telegram bind failed", slog.Any("error", err))
		b.sendText(msg.Chat.ID, "❌ Код недействителен или истёк\\.")

		return
	}

	b.sendText(msg.Chat.ID, "✅ Аккаунт PushPost успешно привязан\\!")
}

func (b *Bot) sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	if _, err := b.api.Send(msg); err != nil {
		b.log.Warn("failed to send telegram message",
			slog.Int64("chat_id", chatID),
			slog.Any("error", err),
		)
	}
}
