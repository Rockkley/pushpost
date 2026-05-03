package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/database"
	"github.com/rockkley/pushpost/services/profile_service/internal/domain"
	"github.com/rockkley/pushpost/services/profile_service/internal/entity"
)

type ProfileRepository struct {
	exec database.Executor
}

func NewProfileRepository(exec database.Executor) *ProfileRepository {
	return &ProfileRepository{exec: exec}
}

func (r *ProfileRepository) Create(ctx context.Context, profile *entity.Profile) error {
	query := `
		INSERT INTO profiles (user_id, username, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (user_id) DO NOTHING`

	_, err := r.exec.ExecContext(ctx, query, profile.UserID, profile.Username)
	if err != nil {
		return fmt.Errorf("profile create: %w", err)
	}
	return nil
}

func (r *ProfileRepository) FindByUsername(ctx context.Context, username string) (*entity.Profile, error) {
	username = strings.ToLower(strings.TrimSpace(username))

	query := `
		SELECT user_id, username, display_name, first_name, last_name, birth_date, avatar_url, bio, telegram_link,
		       is_private, created_at, updated_at, deleted_at
		FROM   profiles
		WHERE  LOWER(username) = $1
		  AND  deleted_at IS NULL`

	return r.scanProfile(r.exec.QueryRowContext(ctx, query, username))
}

func (r *ProfileRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.Profile, error) {
	query := `
		SELECT user_id, username, display_name, first_name, last_name, birth_date, avatar_url, bio, telegram_link,
		       is_private, created_at, updated_at, deleted_at
		FROM   profiles
		WHERE  user_id = $1
		  AND  deleted_at IS NULL`

	return r.scanProfile(r.exec.QueryRowContext(ctx, query, userID))
}

func (r *ProfileRepository) Update(ctx context.Context, profile *entity.Profile) error {
	query := `
		UPDATE profiles
		SET    display_name = $2,
		       first_name = $3,
		       last_name = $4,
		       birth_date = $5,
		       avatar_url = $6,
		       bio = $7,
		       telegram_link = $8,
		       is_private = $9
		WHERE  user_id = $1
		  AND  deleted_at IS NULL`

	res, err := r.exec.ExecContext(
		ctx, query,
		profile.UserID,
		profile.DisplayName,
		profile.FirstName,
		profile.LastName,
		profile.BirthDate,
		profile.AvatarURL,
		profile.Bio,
		profile.TelegramLink,
		profile.IsPrivate,
	)
	if err != nil {
		return commonapperr.MapPostgresError(err, "update profile")
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update profile rows affected: %w", err)
	}
	if affected == 0 {
		return domain.ErrProfileNotFound
	}
	return nil
}

func (r *ProfileRepository) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	query := `
		UPDATE profiles
		SET    avatar_url = $2
		WHERE  user_id = $1
		  AND  deleted_at IS NULL`

	res, err := r.exec.ExecContext(ctx, query, userID, avatarURL)
	if err != nil {
		return commonapperr.MapPostgresError(err, "update avatar")
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update avatar rows affected: %w", err)
	}
	if affected == 0 {
		return domain.ErrProfileNotFound
	}
	return nil
}

func (r *ProfileRepository) scanProfile(row *sql.Row) (*entity.Profile, error) {
	var p entity.Profile
	err := row.Scan(
		&p.UserID,
		&p.Username,
		&p.DisplayName,
		&p.FirstName,
		&p.LastName,
		&p.BirthDate,
		&p.AvatarURL,
		&p.Bio,
		&p.TelegramLink,
		&p.IsPrivate,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProfileNotFound
		}
		return nil, commonapperr.MapPostgresError(err, "scan profile")
	}
	return &p, nil
}
