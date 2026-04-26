package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/database"
	"github.com/rockkley/pushpost/services/notification_service/internal/entity"
	"github.com/rockkley/pushpost/services/notification_service/internal/repository"
)

type telegramRepo struct{ exec database.Executor }

func NewTelegramRepository(exec database.Executor) repository.TelegramRepository {
	return &telegramRepo{exec: exec}
}

func (r *telegramRepo) Upsert(ctx context.Context, b *entity.TelegramBinding) error {
	const q = `INSERT INTO telegram_bindings (user_id, chat_id, username, created_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (user_id) DO UPDATE SET chat_id = EXCLUDED.chat_id, username = EXCLUDED.username`
	_, err := r.exec.ExecContext(ctx, q, b.UserID, b.ChatID, b.Username)
	if err != nil {
		return commonapperr.MapPostgresError(err, "upsert telegram binding")
	}
	return nil
}

func (r *telegramRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.TelegramBinding, error) {
	const q = `SELECT user_id, chat_id, username, created_at FROM telegram_bindings WHERE user_id = $1`
	var b entity.TelegramBinding
	err := r.exec.QueryRowContext(ctx, q, userID).Scan(&b.UserID, &b.ChatID, &b.Username, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, commonapperr.MapPostgresError(err, "find telegram binding")
	}
	return &b, nil
}

func (r *telegramRepo) Delete(ctx context.Context, userID uuid.UUID) error {
	const q = `DELETE FROM telegram_bindings WHERE user_id = $1`
	_, err := r.exec.ExecContext(ctx, q, userID)
	if err != nil {
		return commonapperr.MapPostgresError(err, "delete telegram binding")
	}
	return nil
}
