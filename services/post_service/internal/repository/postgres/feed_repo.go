package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/database"
	"github.com/rockkley/pushpost/services/post_service/internal/entity"
)

type FeedRepository struct {
	exec database.Executor
}

func NewFeedRepository(exec database.Executor) *FeedRepository {
	return &FeedRepository{exec: exec}
}

func (r *FeedRepository) InsertBatch(
	ctx context.Context,
	postID uuid.UUID,
	userIDs []uuid.UUID,
	insertedAt time.Time,
) error {
	if len(userIDs) == 0 {
		return nil
	}
	slog.Info("inserting feed", slog.Int("users", len(userIDs)))
	// Один INSERT со всеми получателями.
	// Параметры: $1..$N — user_id, $(N+1) — post_id, $(N+2) — inserted_at
	valueStrings := make([]string, len(userIDs))
	args := make([]any, 0, len(userIDs)+2)

	for i, userID := range userIDs {
		valueStrings[i] = fmt.Sprintf("($%d, $%d, $%d)", i+1, len(userIDs)+1, len(userIDs)+2)
		args = append(args, userID)
	}

	args = append(args, postID, insertedAt)

	query := fmt.Sprintf(
		`INSERT INTO feeds (user_id, post_id, inserted_at) VALUES %s ON CONFLICT DO NOTHING`,
		strings.Join(valueStrings, ","),
	)

	_, err := r.exec.ExecContext(ctx, query, args...)

	if err != nil {
		return commonapperr.MapPostgresError(err, "feed insert batch")
	}

	return nil
}

func (r *FeedRepository) GetFeed(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
	before time.Time,
	beforeID uuid.UUID,
) ([]*entity.Post, error) {
	query := `
		SELECT p.id, p.author_id, p.content, p.version, p.likes_count, p.dislikes_count,
		       p.likes_count - p.dislikes_count AS rating, p.created_at, p.updated_at, f.inserted_at

		FROM feeds f
		JOIN posts p ON p.id = f.post_id
		WHERE f.user_id = $1
		  AND p.deleted_at IS NULL
		  AND (f.inserted_at, f.post_id) < ($2, $3)
		ORDER BY f.inserted_at DESC, f.post_id DESC
		LIMIT $4`

	rows, err := r.exec.QueryContext(ctx, query, userID, before, beforeID, limit)

	if err != nil {
		return nil, commonapperr.MapPostgresError(err, "get feed")
	}

	defer rows.Close()

	return scanFeedPosts(rows)
}

func (r *FeedRepository) GetFeedSince(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
	after time.Time,
	afterID uuid.UUID,
) ([]*entity.Post, error) {
	// Возвращает посты НОВЕЕ курсора (для reconciliation / refresh)
	// Сортировка ASC — потом разворачиваем на уровне usecase
	query := `
		SELECT p.id, p.author_id, p.content, p.version, p.likes_count, p.dislikes_count,
		       p.likes_count - p.dislikes_count AS rating, p.created_at, p.updated_at, f.inserted_at

		FROM feeds f
		JOIN posts p ON p.id = f.post_id
		WHERE f.user_id = $1
		  AND p.deleted_at IS NULL
		  AND (f.inserted_at, f.post_id) > ($2, $3)
		ORDER BY f.inserted_at ASC, f.post_id ASC
		LIMIT $4`

	rows, err := r.exec.QueryContext(ctx, query, userID, after, afterID, limit)

	if err != nil {
		return nil, commonapperr.MapPostgresError(err, "get feed since")
	}

	defer rows.Close()

	posts, err := scanFeedPosts(rows)

	if err != nil {
		return nil, err
	}

	// Разворачиваем — клиент ожидает DESC (новые сверху)
	for i, j := 0, len(posts)-1; i < j; i, j = i+1, j-1 {
		posts[i], posts[j] = posts[j], posts[i]
	}

	return posts, nil
}

func (r *FeedRepository) FindRecipients(ctx context.Context, postID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT user_id FROM feeds WHERE post_id = $1`

	rows, err := r.exec.QueryContext(ctx, query, postID)

	if err != nil {
		return nil, commonapperr.MapPostgresError(err, "find feed recipients")
	}

	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID

		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

func (r *FeedRepository) DeleteByPostID(ctx context.Context, postID uuid.UUID) error {
	_, err := r.exec.ExecContext(ctx, `DELETE FROM feeds WHERE post_id = $1`, postID)
	if err != nil {
		return commonapperr.MapPostgresError(err, "feed delete by post")
	}
	return nil
}

func (r *FeedRepository) DeleteByAuthor(ctx context.Context, recipientID, authorID uuid.UUID) error {
	query := `
		DELETE FROM feeds
		WHERE user_id = $1
		  AND post_id IN (SELECT id FROM posts WHERE author_id = $2)`

	_, err := r.exec.ExecContext(ctx, query, recipientID, authorID)

	if err != nil {
		return commonapperr.MapPostgresError(err, "feed delete by author")
	}

	return nil
}

func (r *FeedRepository) DeleteUserFeed(ctx context.Context, userID uuid.UUID) error {
	_, err := r.exec.ExecContext(ctx, `DELETE FROM feeds WHERE user_id = $1`, userID)

	if err != nil {
		return commonapperr.MapPostgresError(err, "delete user feed")
	}

	return nil
}

func (r *FeedRepository) DeleteByAuthorFromAllFeeds(ctx context.Context, authorID uuid.UUID) error {
	query := `
		DELETE FROM feeds
		WHERE post_id IN (SELECT id FROM posts WHERE author_id = $1)`

	_, err := r.exec.ExecContext(ctx, query, authorID)

	if err != nil {
		return commonapperr.MapPostgresError(err, "delete author posts from all feeds")
	}

	return nil
}

func scanFeedPosts(rows *sql.Rows) ([]*entity.Post, error) {
	var result []*entity.Post

	for rows.Next() {
		var p entity.Post
		if err := rows.Scan(
			&p.ID, &p.AuthorID, &p.Content, &p.Version,
			&p.LikesCount, &p.DislikesCount, &p.Rating,
			&p.CreatedAt, &p.UpdatedAt,
			&p.InsertedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, &p)
	}

	return result, rows.Err()
}
