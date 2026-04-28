package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/database"
	apperr "github.com/rockkley/pushpost/services/notification_service/internal/apperror"
	"github.com/rockkley/pushpost/services/notification_service/internal/entity"
	"github.com/rockkley/pushpost/services/notification_service/internal/repository"
)

type notificationRepo struct{ exec database.Executor }

func NewNotificationRepository(exec database.Executor) repository.NotificationRepository {
	return &notificationRepo{exec: exec}
}

func (r *notificationRepo) Create(ctx context.Context, n *entity.Notification) error {
	data, err := json.Marshal(n.Data)

	if err != nil {
		return fmt.Errorf("marshal notification data: %w", err)
	}

	query := `
		INSERT INTO notifications (id, user_id, type, title, body, data, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (id) DO NOTHING
		RETURNING created_at`

	err = r.exec.QueryRowContext(ctx, query,
		n.ID, n.UserID, string(n.Type), n.Title, n.Body, data,
	).Scan(&n.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return commonapperr.MapPostgresError(err, "create notification")
	}
	return nil
}

func (r *notificationRepo) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, error) {
	query := `
		SELECT id, user_id, type, title, body, data, read_at, created_at
		FROM   notifications
		WHERE  user_id = $1
		ORDER  BY created_at DESC
		LIMIT  $2 OFFSET $3`

	rows, err := r.exec.QueryContext(ctx, query, userID, limit, offset)

	if err != nil {
		return nil, commonapperr.MapPostgresError(err, "get notifications")
	}

	defer rows.Close()
	return scanNotifications(rows)
}

func (r *notificationRepo) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	const q = `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`
	var count int

	if err := r.exec.QueryRowContext(ctx, q, userID).Scan(&count); err != nil {
		return 0, commonapperr.MapPostgresError(err, "get unread count")
	}

	return count, nil
}

func (r *notificationRepo) MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	updateQ := `
		UPDATE notifications
		SET    read_at = $1
		WHERE  id = $2 AND user_id = $3 AND read_at IS NULL`

	result, err := r.exec.ExecContext(ctx, updateQ, time.Now().UTC(), notificationID, userID)

	if err != nil {
		return commonapperr.MapPostgresError(err, "mark notification as read")
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return commonapperr.Internal("mark as read rows affected", err)
	}

	if rows > 0 {
		return nil
	}

	existsQ := `
		SELECT EXISTS(
			SELECT 1 FROM notifications WHERE id = $1 AND user_id = $2
		)`

	var exists bool

	if err = r.exec.QueryRowContext(ctx, existsQ, notificationID, userID).Scan(&exists); err != nil {
		return commonapperr.MapPostgresError(err, "check notification existence")
	}

	if !exists {
		return apperr.NotificationNotFound()
	}

	return nil
}

func (r *notificationRepo) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	const q = `UPDATE notifications SET read_at = $1 WHERE user_id = $2 AND read_at IS NULL`
	_, err := r.exec.ExecContext(ctx, q, time.Now().UTC(), userID)

	if err != nil {
		return commonapperr.MapPostgresError(err, "mark all notifications as read")
	}

	return nil
}

func scanNotifications(rows *sql.Rows) ([]*entity.Notification, error) {
	result := make([]*entity.Notification, 0)

	for rows.Next() {
		var n entity.Notification
		var rawData []byte

		if err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body,
			&rawData, &n.ReadAt, &n.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}

		if len(rawData) > 0 {
			if err := json.Unmarshal(rawData, &n.Data); err != nil {
				return nil, fmt.Errorf("unmarshal notification data: %w", err)
			}
		}
		result = append(result, &n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return result, nil
}
