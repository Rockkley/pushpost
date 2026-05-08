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

type CommentRepository struct{ exec database.Executor }

func NewCommentRepository(exec database.Executor) repository.CommentRepositoryInterface {
	return &CommentRepository{exec: exec}
}

func (r *CommentRepository) CreateComment(ctx context.Context, comment *entity.Comment) error {
	var parentID interface{}
	if comment.ParentID != nil {
		parentID = *comment.ParentID
	}

	query := `
		WITH parent_comment AS (
			SELECT c.id, c.author_id FROM comments c WHERE c.id = $3 AND c.post_id = $2
		)
		INSERT INTO comments (id, post_id, author_id, parent_id, reply_to_user_id, content)
		SELECT $1, p.id, $4,
		       (SELECT id FROM parent_comment),
		       (SELECT author_id FROM parent_comment),
		       $5
		FROM posts p
		WHERE p.id = $2
		  AND p.deleted_at IS NULL
		  AND ($3::uuid IS NULL OR EXISTS (SELECT 1 FROM parent_comment))
		RETURNING post_id, parent_id, reply_to_user_id, content, upvotes_count, downvotes_count,
		          upvotes_count-downvotes_count AS rating, created_at, updated_at`
	err := r.exec.QueryRowContext(ctx, query, comment.ID, comment.PostID, parentID, comment.AuthorID, comment.Content).
		Scan(&comment.PostID, &comment.ParentID, &comment.ReplyToUserID, &comment.Content, &comment.UpvotesCount, &comment.DownvotesCount, &comment.Rating, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.PostNotFound()
		}
		return commonapperr.MapPostgresError(err, "create comment")
	}
	return nil
}

func (r *CommentRepository) GetCommentsByPostID(ctx context.Context, postID uuid.UUID, limit int, after time.Time, afterID uuid.UUID) ([]*entity.Comment, error) {
	query := `SELECT c.id, c.post_id, c.author_id, c.parent_id, c.reply_to_user_id, c.content,
		c.upvotes_count, c.downvotes_count, c.upvotes_count-c.downvotes_count AS rating,
		c.created_at, c.updated_at FROM comments c JOIN posts p ON p.id = c.post_id
		WHERE c.post_id=$1 AND p.deleted_at IS NULL AND (c.created_at, c.id) > ($2, $3)
		ORDER BY c.created_at ASC, c.id ASC LIMIT $4`
	rows, err := r.exec.QueryContext(ctx, query, postID, after, afterID, limit)
	if err != nil {
		return nil, commonapperr.MapPostgresError(err, "get comments by post id")
	}
	defer rows.Close()
	var result []*entity.Comment
	for rows.Next() {
		c, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (r *CommentRepository) FindCommentByID(ctx context.Context, commentID uuid.UUID) (*entity.Comment, error) {
	q := `SELECT id, post_id, author_id, parent_id, reply_to_user_id, content, upvotes_count, downvotes_count,
	      upvotes_count-downvotes_count AS rating, created_at, updated_at FROM comments WHERE id=$1`
	var c entity.Comment
	err := r.exec.QueryRowContext(ctx, q, commentID).Scan(&c.ID, &c.PostID, &c.AuthorID, &c.ParentID, &c.ReplyToUserID, &c.Content, &c.UpvotesCount, &c.DownvotesCount, &c.Rating, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.CommentNotFound()
		}
		return nil, commonapperr.MapPostgresError(err, "find comment")
	}
	return &c, nil
}

func (r *CommentRepository) UpdateComment(ctx context.Context, comment *entity.Comment) error {
	q := `UPDATE comments SET content=$1 WHERE id=$2 AND author_id=$3 RETURNING post_id,parent_id,reply_to_user_id,upvotes_count,downvotes_count,upvotes_count-downvotes_count AS rating,created_at,updated_at`
	err := r.exec.QueryRowContext(ctx, q, comment.Content, comment.ID, comment.AuthorID).Scan(&comment.PostID, &comment.ParentID, &comment.ReplyToUserID, &comment.UpvotesCount, &comment.DownvotesCount, &comment.Rating, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.CommentNotFound()
		}
		return commonapperr.MapPostgresError(err, "update comment")
	}
	return nil
}

func (r *CommentRepository) SetCommentVote(ctx context.Context, commentID, userID uuid.UUID, value int) (*entity.Comment, error) {
	if value != 1 && value != -1 {
		return nil, commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid vote value")
	}
	query := `
		WITH target AS (
			SELECT id FROM comments WHERE id = $1
		),
		existing AS (
			SELECT value FROM comment_votes WHERE comment_id = $1 AND user_id = $2
		),
		upserted AS (
			INSERT INTO comment_votes (comment_id, user_id, value)
			SELECT id, $2, $3 FROM target
			ON CONFLICT (comment_id, user_id) DO UPDATE SET value = EXCLUDED.value
			RETURNING value
		),
		delta AS (
			SELECT COALESCE((SELECT value FROM existing), 0) AS old_value,
			       COALESCE((SELECT value FROM upserted), 0) AS new_value
		)
		UPDATE comments c
		SET upvotes_count = c.upvotes_count
			+ CASE WHEN (SELECT old_value FROM delta) = 1 THEN -1 ELSE 0 END
			+ CASE WHEN (SELECT new_value FROM delta) = 1 THEN 1 ELSE 0 END,
		    downvotes_count = c.downvotes_count
			+ CASE WHEN (SELECT old_value FROM delta) = -1 THEN -1 ELSE 0 END
			+ CASE WHEN (SELECT new_value FROM delta) = -1 THEN 1 ELSE 0 END
		WHERE c.id IN (SELECT id FROM target)
		RETURNING c.id, c.post_id, c.author_id, c.parent_id, c.reply_to_user_id, c.content,
		          c.upvotes_count, c.downvotes_count, c.upvotes_count - c.downvotes_count AS rating,
		          c.created_at, c.updated_at`
	var c entity.Comment
	err := r.exec.QueryRowContext(ctx, query, commentID, userID, value).Scan(
		&c.ID, &c.PostID, &c.AuthorID, &c.ParentID, &c.ReplyToUserID, &c.Content,
		&c.UpvotesCount, &c.DownvotesCount, &c.Rating, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.CommentNotFound()
		}
		return nil, commonapperr.MapPostgresError(err, "vote comment")
	}
	return &c, nil
}

func (r *CommentRepository) RemoveCommentVote(ctx context.Context, commentID, userID uuid.UUID) (*entity.Comment, error) {
	query := `
		WITH target AS (
			SELECT id FROM comments WHERE id = $1
		),
		removed AS (
			DELETE FROM comment_votes
			WHERE comment_id = $1 AND user_id = $2
			RETURNING value
		),
		delta AS (
			SELECT COALESCE((SELECT value FROM removed), 0) AS old_value
		)
		UPDATE comments c
		SET upvotes_count = c.upvotes_count
			+ CASE WHEN (SELECT old_value FROM delta) = 1 THEN -1 ELSE 0 END,
		    downvotes_count = c.downvotes_count
			+ CASE WHEN (SELECT old_value FROM delta) = -1 THEN -1 ELSE 0 END
		WHERE c.id IN (SELECT id FROM target)
		RETURNING c.id, c.post_id, c.author_id, c.parent_id, c.reply_to_user_id, c.content,
		          c.upvotes_count, c.downvotes_count, c.upvotes_count - c.downvotes_count AS rating,
		          c.created_at, c.updated_at`
	var c entity.Comment
	err := r.exec.QueryRowContext(ctx, query, commentID, userID).Scan(
		&c.ID, &c.PostID, &c.AuthorID, &c.ParentID, &c.ReplyToUserID, &c.Content,
		&c.UpvotesCount, &c.DownvotesCount, &c.Rating, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.CommentNotFound()
		}
		return nil, commonapperr.MapPostgresError(err, "remove comment vote")
	}
	return &c, nil
}

func (r *CommentRepository) DeleteComment(ctx context.Context, commentID, authorID uuid.UUID) error {
	res, err := r.exec.ExecContext(ctx, `DELETE FROM comments WHERE id = $1 AND author_id = $2`, commentID, authorID)

	if err != nil {
		return commonapperr.MapPostgresError(err, "delete comment")
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return commonapperr.Internal("rows affected", err)
	}

	if rows == 0 {
		return apperror.CommentNotFound()
	}

	return nil
}

func scanComment(rows *sql.Rows) (*entity.Comment, error) {
	var c entity.Comment
	if err := rows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.ParentID, &c.ReplyToUserID, &c.Content, &c.UpvotesCount, &c.DownvotesCount, &c.Rating, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, commonapperr.Internal("scan comment", err)
	}

	return &c, nil
}
