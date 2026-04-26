package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/rockkley/pushpost/services/notification_service/internal/entity"
	"github.com/rockkley/pushpost/services/notification_service/internal/repository"
)

type MessageSender interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

// Deliverer sends a notification via Telegram.
// If the user has no Telegram binding, Deliver is a no-op (returns nil).
type Deliverer struct {
	telegramRepo repository.TelegramRepository
	sender       MessageSender
}

func NewDeliverer(repo repository.TelegramRepository, sender MessageSender) *Deliverer {
	return &Deliverer{telegramRepo: repo, sender: sender}
}

func (d *Deliverer) Channel() entity.Channel { return entity.ChannelTelegram }

func (d *Deliverer) Deliver(ctx context.Context, n *entity.Notification) error {
	binding, err := d.telegramRepo.FindByUserID(ctx, n.UserID)
	if err != nil {
		return fmt.Errorf("find telegram binding for user %s: %w", n.UserID, err)
	}
	if binding == nil {
		return nil
	}

	text := fmt.Sprintf("🔔 *%s*\n%s", escapeMarkdownV2(n.Title), escapeMarkdownV2(n.Body))
	return d.sender.SendMessage(ctx, binding.ChatID, text)
}

func escapeMarkdownV2(s string) string {
	return strings.NewReplacer(
		`_`, `\_`, `*`, `\*`, `[`, `\[`, `]`, `\]`,
		`(`, `\(`, `)`, `\)`, `~`, `\~`, "`", "\\`",
		`>`, `\>`, `#`, `\#`, `+`, `\+`, `-`, `\-`,
		`=`, `\=`, `|`, `\|`, `{`, `\{`, `}`, `\}`,
		`.`, `\.`, `!`, `\!`,
	).Replace(s)
}
