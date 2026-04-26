package postgres

import (
	"context"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/database"
	"github.com/rockkley/pushpost/services/notification_service/internal/entity"
	"github.com/rockkley/pushpost/services/notification_service/internal/repository"
)

type preferenceRepo struct {
	exec database.Executor
}

func NewPreferenceRepository(exec database.Executor) repository.PreferenceRepository {
	return &preferenceRepo{exec: exec}
}

func (r *preferenceRepo) GetAll(ctx context.Context, userID uuid.UUID) ([]*entity.NotificationPreference, error) {
	const query = `
		SELECT user_id, type, channel, enabled
		FROM notification_preferences
		WHERE user_id = $1
		ORDER BY type, channel`

	rows, err := r.exec.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, commonapperr.MapPostgresError(err, "get all preferences")
	}
	defer rows.Close()

	var result []*entity.NotificationPreference
	for rows.Next() {
		var p entity.NotificationPreference
		if err = rows.Scan(&p.UserID, &p.Type, &p.Channel, &p.Enabled); err != nil {
			return nil, commonapperr.MapPostgresError(err, "scan preference")
		}
		result = append(result, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, commonapperr.MapPostgresError(err, "iterate preferences")
	}
	return result, nil
}

func (r *preferenceRepo) GetEnabled(ctx context.Context, userID uuid.UUID, nType entity.NotificationType) ([]entity.Channel, error) {
	const query = `
		SELECT channel, enabled
		FROM notification_preferences
		WHERE user_id = $1 AND type = $2`

	rows, err := r.exec.QueryContext(ctx, query, userID, string(nType))
	if err != nil {
		return nil, commonapperr.MapPostgresError(err, "get enabled channels")
	}
	defer rows.Close()

	var (
		hasPrefs        bool
		enabledChannels []entity.Channel
	)

	for rows.Next() {
		var ch entity.Channel
		var enabled bool
		if err = rows.Scan(&ch, &enabled); err != nil {
			return nil, commonapperr.MapPostgresError(err, "scan channel preference")
		}
		hasPrefs = true
		if enabled {
			enabledChannels = append(enabledChannels, ch)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, commonapperr.MapPostgresError(err, "iterate channel preferences")
	}

	if !hasPrefs {
		return entity.AllChannels, nil
	}
	return enabledChannels, nil
}

func (r *preferenceRepo) Upsert(ctx context.Context, pref *entity.NotificationPreference) error {
	const query = `
		INSERT INTO notification_preferences (user_id, type, channel, enabled, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, type, channel)
		DO UPDATE SET enabled = EXCLUDED.enabled, updated_at = NOW()`

	_, err := r.exec.ExecContext(ctx, query,
		pref.UserID, string(pref.Type), string(pref.Channel), pref.Enabled,
	)
	if err != nil {
		return commonapperr.MapPostgresError(err, "upsert preference")
	}
	return nil
}
