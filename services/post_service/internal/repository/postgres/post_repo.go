package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/database"
	"github.com/rockkley/pushpost/services/post_service/internal/apperror"
	"github.com/rockkley/pushpost/services/post_service/internal/entity"
	"github.com/rockkley/pushpost/services/post_service/internal/repository"
)

type PostRepository struct {
	exec database.Executor
}

func NewPostRepository(exec database.Executor) repository.PostRepositoryInterface {
	return &PostRepository{exec: exec}
}

func (r *PostRepository) Create(ctx context.Context, post *entity.Post) error {
	query := `
		INSERT INTO posts (id, author_id, content)
		VALUES ($1, $2, $3)
		RETURNING version, created_at, updated_at`

	err := r.exec.QueryRowContext(ctx, query, post.ID, post.AuthorID, post.Content).
		Scan(&post.Version, &post.CreatedAt, &post.UpdatedAt)

	if err != nil {
		return commonapperr.MapPostgresError(err, "create post")
	}

	return nil
}

func (r *PostRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Post, error) {
	query := `
		SELECT id, author_id, content, version, likes_count, dislikes_count,
		       likes_count - dislikes_count AS rating,
		       created_at, updated_at, deleted_at

		FROM posts WHERE id = $1`

	var p entity.Post
	err := r.exec.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.AuthorID, &p.Content, &p.Version,
		&p.LikesCount, &p.DislikesCount, &p.Rating,
		&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.PostNotFound()
		}

		return nil, commonapperr.MapPostgresError(err, "find post by id")
	}

	return &p, nil
}

func (r *PostRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Post, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `
		SELECT id, author_id, content, version, likes_count, dislikes_count,
		       likes_count - dislikes_count AS rating, created_at, updated_at

		FROM posts
		WHERE id = ANY($1::uuid[])
		  AND deleted_at IS NULL`

	rows, err := r.exec.QueryContext(ctx, query, ids)
	if err != nil {
		return nil, commonapperr.MapPostgresError(err, "get posts by ids")
	}
	defer rows.Close()

	return scanPosts(rows)
}

func (r *PostRepository) Update(ctx context.Context, post *entity.Post) error {
	query := `
		UPDATE posts
		SET content = $1,
		    version = version + 1
		WHERE id = $2
		  AND author_id = $3
		  AND deleted_at IS NULL
		RETURNING version, updated_at`

	err := r.exec.QueryRowContext(ctx, query, post.Content, post.ID, post.AuthorID).
		Scan(&post.Version, &post.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.PostNotFound()
		}

		return commonapperr.MapPostgresError(err, "update post")
	}

	return nil
}

func (r *PostRepository) GetByAuthors(
	ctx context.Context,
	authorIDs []uuid.UUID,
	limit int,
	before time.Time,
	beforeID uuid.UUID,
) ([]*entity.Post, error) {
	if len(authorIDs) == 0 {
		return nil, nil
	}

	query := `
		SELECT id, author_id, content, version, likes_count, dislikes_count,
		       likes_count - dislikes_count AS rating, created_at, updated_at
		FROM posts
		WHERE author_id = ANY($1::uuid[])
		  AND deleted_at IS NULL
		  AND (created_at, id) < ($2, $3)
		ORDER BY created_at DESC, id DESC
		LIMIT $4`

	rows, err := r.exec.QueryContext(ctx, query, authorIDs, before, beforeID, limit)
	if err != nil {
		return nil, commonapperr.MapPostgresError(err, "get posts by authors")
	}
	defer rows.Close()

	return scanPosts(rows)
}

func (r *PostRepository) GetByAuthor(
	ctx context.Context,
	authorID uuid.UUID,
	limit int,
	before time.Time,
	beforeID uuid.UUID,
) ([]*entity.Post, error) {
	query := `
		SELECT id, author_id, content, version, likes_count, dislikes_count,
		       likes_count - dislikes_count AS rating, created_at, updated_at
		FROM posts
		WHERE author_id = $1
		  AND deleted_at IS NULL
		  AND (created_at, id) < ($2, $3)
		ORDER BY created_at DESC, id DESC
		LIMIT $4`

	rows, err := r.exec.QueryContext(ctx, query, authorID, before, beforeID, limit)

	if err != nil {
		return nil, commonapperr.MapPostgresError(err, "get posts by author")
	}
	defer rows.Close()

	return scanPosts(rows)
}

func (r *PostRepository) SetVote(ctx context.Context, postID, userID uuid.UUID, value int) (*entity.Post, error) {
	query := `
		WITH target AS (
			SELECT id FROM posts WHERE id = $1 AND deleted_at IS NULL
		),
		existing AS (
			SELECT pv.value
			FROM post_votes pv
			JOIN target t ON t.id = pv.post_id
			WHERE pv.user_id = $2
		),
		upsert AS (
			INSERT INTO post_votes (post_id, user_id, value)
			SELECT t.id, $2, $3 FROM target t
			ON CONFLICT (post_id, user_id) DO UPDATE SET value = EXCLUDED.value
			RETURNING value
		),
		delta AS (
			SELECT COALESCE((SELECT value FROM existing), 0) AS old_value,
			       COALESCE((SELECT value FROM upsert), 0) AS new_value
		)
		UPDATE posts p
		SET likes_count = p.likes_count
			+ CASE WHEN (SELECT old_value FROM delta) = 1 THEN -1 ELSE 0 END
			+ CASE WHEN (SELECT new_value FROM delta) = 1 THEN 1 ELSE 0 END,
		    dislikes_count = p.dislikes_count
			+ CASE WHEN (SELECT old_value FROM delta) = -1 THEN -1 ELSE 0 END
			+ CASE WHEN (SELECT new_value FROM delta) = -1 THEN 1 ELSE 0 END
		WHERE p.id IN (SELECT id FROM target)
		RETURNING p.id, p.author_id, p.content, p.version, p.likes_count, p.dislikes_count,
		          p.likes_count - p.dislikes_count AS rating, p.created_at, p.updated_at`

	var post entity.Post

	err := r.exec.QueryRowContext(ctx, query, postID, userID, value).Scan(
		&post.ID, &post.AuthorID, &post.Content, &post.Version,
		&post.LikesCount, &post.DislikesCount, &post.Rating,
		&post.CreatedAt, &post.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.PostNotFound()
		}
		return nil, commonapperr.MapPostgresError(err, "set post vote")
	}

	return &post, nil
}

func (r *PostRepository) RemoveVote(ctx context.Context, postID, userID uuid.UUID) (*entity.Post, error) {
	query := `
		WITH target AS (
			SELECT id FROM posts WHERE id = $1 AND deleted_at IS NULL
		),
		removed AS (
			DELETE FROM post_votes pv
			USING target t
			WHERE pv.post_id = t.id
			  AND pv.user_id = $2
			RETURNING pv.value
		),
		delta AS (
			SELECT COALESCE((SELECT value FROM removed), 0) AS old_value
		)
		UPDATE posts p
		SET likes_count = p.likes_count
			+ CASE WHEN (SELECT old_value FROM delta) = 1 THEN -1 ELSE 0 END,
		    dislikes_count = p.dislikes_count
			+ CASE WHEN (SELECT old_value FROM delta) = -1 THEN -1 ELSE 0 END
		WHERE p.id IN (SELECT id FROM target)
		RETURNING p.id, p.author_id, p.content, p.version, p.likes_count, p.dislikes_count,
		          p.likes_count - p.dislikes_count AS rating, p.created_at, p.updated_at`

	var post entity.Post

	err := r.exec.QueryRowContext(ctx, query, postID, userID).Scan(
		&post.ID, &post.AuthorID, &post.Content, &post.Version,
		&post.LikesCount, &post.DislikesCount, &post.Rating,
		&post.CreatedAt, &post.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.PostNotFound()
		}

		return nil, commonapperr.MapPostgresError(err, "remove post vote")
	}

	return &post, nil
}

func (r *PostRepository) SoftDelete(ctx context.Context, postID, authorID uuid.UUID) error {
	query := `
		UPDATE posts SET deleted_at = NOW()
		WHERE id = $1 AND author_id = $2 AND deleted_at IS NULL`

	result, err := r.exec.ExecContext(ctx, query, postID, authorID)

	if err != nil {
		return commonapperr.MapPostgresError(err, "soft delete post")
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return commonapperr.Internal("rows affected", err)
	}

	if rows == 0 {
		return apperror.PostNotFound()
	}

	return nil
}

func scanPosts(rows *sql.Rows) ([]*entity.Post, error) {
	var result []*entity.Post
	for rows.Next() {
		var p entity.Post
		if err := rows.Scan(
			&p.ID, &p.AuthorID, &p.Content, &p.Version,
			&p.LikesCount, &p.DislikesCount, &p.Rating,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, &p)
	}

	return result, rows.Err()
}
