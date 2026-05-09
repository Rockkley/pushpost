package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/rockkley/pushpost/services/profile_service/internal/domain/dto"
	"strconv"
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
		SELECT user_id, username, display_name, first_name, last_name, city, country, birth_date, avatar_url, avatar_thumb_url, bio, telegram_link, github_link,
		       is_private, created_at, updated_at, deleted_at
		FROM   profiles
		WHERE  LOWER(username) = $1
		  AND  deleted_at IS NULL`

	return r.scanProfile(r.exec.QueryRowContext(ctx, query, username))
}

func (r *ProfileRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.Profile, error) {
	query := `
		SELECT user_id, username, display_name, first_name, last_name, city, country, birth_date, avatar_url, avatar_thumb_url, bio, telegram_link, github_link,
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
		       city = $5,
		       country = $6,
		       birth_date = $7,
		       avatar_url = $8,
		       bio = $9,
		       telegram_link = $10,
		       github_link = $11,
		       is_private = $12

		WHERE  user_id = $1
		  AND  deleted_at IS NULL`

	res, err := r.exec.ExecContext(
		ctx, query,
		profile.UserID,
		profile.DisplayName,
		profile.FirstName,
		profile.LastName,
		profile.City,
		profile.Country,
		profile.BirthDate,
		profile.AvatarURL,
		profile.Bio,
		profile.TelegramLink,
		profile.GithubLink,
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

func (r *ProfileRepository) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string, avatarThumbURL string) error {
	query := `
		UPDATE profiles
		SET    avatar_url = $2, avatar_thumb_url = $3
		WHERE  user_id = $1
		  AND  deleted_at IS NULL`

	res, err := r.exec.ExecContext(ctx, query, userID, avatarURL, avatarThumbURL)

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
		&p.City,
		&p.Country,
		&p.BirthDate,
		&p.AvatarURL,
		&p.AvatarThumbURL,
		&p.Bio,
		&p.TelegramLink,
		&p.GithubLink,
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

func (r *ProfileRepository) Search(ctx context.Context, filter *dto.SearchProfilesQuery) ([]*entity.Profile, error) {
	if filter == nil {
		filter = &dto.SearchProfilesQuery{}
	}

	filter.NormalizePagination()

	conditions := []string{"deleted_at IS NULL"}
	args := make([]interface{}, 0, 4)
	idx := 1

	username := strings.TrimSpace(filter.Username)

	if username != "" {
		conditions = append(conditions, "LOWER(username) LIKE LOWER($"+strconv.Itoa(idx)+")")
		args = append(args, "%"+username+"%")
		idx++
	}

	firstName := strings.TrimSpace(filter.FirstName)

	if firstName != "" {
		conditions = append(conditions, "LOWER(first_name) LIKE LOWER($"+strconv.Itoa(idx)+")")
		args = append(args, "%"+firstName+"%")
		idx++
	}

	lastName := strings.TrimSpace(filter.LastName)

	if lastName != "" {
		conditions = append(conditions, "LOWER(last_name) LIKE LOWER($"+strconv.Itoa(idx)+")")
		args = append(args, "%"+lastName+"%")
		idx++
	}

	fullName := strings.TrimSpace(filter.FullName)

	if fullName != "" {
		conditions = append(
			conditions,
			"(LOWER(CONCAT(COALESCE(first_name, ''), ' ', COALESCE(last_name, ''))) LIKE LOWER($"+strconv.Itoa(idx)+"))",
		)
		args = append(args, "%"+fullName+"%")
		idx++
	}

	city := strings.TrimSpace(filter.City)

	if city != "" {
		conditions = append(conditions, "LOWER(city) LIKE LOWER($"+strconv.Itoa(idx)+")")
		args = append(args, "%"+city+"%")
		idx++
	}

	country := strings.TrimSpace(filter.Country)

	if country != "" {
		conditions = append(conditions, "LOWER(country) LIKE LOWER($"+strconv.Itoa(idx)+")")
		args = append(args, "%"+country+"%")
		idx++
	}

	if filter.Age != nil {
		conditions = append(conditions, "birth_date IS NOT NULL")
		conditions = append(conditions, "EXTRACT(YEAR FROM AGE(CURRENT_DATE, birth_date)) = $"+strconv.Itoa(idx))
		args = append(args, *filter.Age)
		idx++
	}

	where := strings.Join(conditions, " AND ")

	query := fmt.Sprintf(`
    SELECT user_id, username, display_name, first_name, last_name, city, country,
           birth_date, avatar_url, avatar_thumb_url, bio, telegram_link, github_link,
           is_private, created_at, updated_at, deleted_at
    FROM profiles
    WHERE %s
    ORDER BY created_at DESC
    LIMIT $%d OFFSET $%d
`, where, idx, idx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.exec.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, commonapperr.MapPostgresError(err, "search profiles")
	}

	defer rows.Close()

	var profiles []*entity.Profile

	for rows.Next() {
		var p entity.Profile
		if err = rows.Scan(
			&p.UserID,
			&p.Username,
			&p.DisplayName,
			&p.FirstName,
			&p.LastName,
			&p.City,
			&p.Country,
			&p.BirthDate,
			&p.AvatarURL,
			&p.AvatarThumbURL,
			&p.Bio,
			&p.TelegramLink,
			&p.GithubLink,
			&p.IsPrivate,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.DeletedAt,
		); err != nil {
			return nil, commonapperr.MapPostgresError(err, "scan profile")
		}

		profiles = append(profiles, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, commonapperr.MapPostgresError(err, "iterate profiles")
	}

	return profiles, nil
}
