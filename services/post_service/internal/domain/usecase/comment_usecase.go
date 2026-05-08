package usecase

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	"github.com/rockkley/pushpost/services/post_service/internal/cursor"
	"github.com/rockkley/pushpost/services/post_service/internal/domain"
	"github.com/rockkley/pushpost/services/post_service/internal/domain/events"
	"github.com/rockkley/pushpost/services/post_service/internal/entity"
)

var mentionRegexp = regexp.MustCompile(`@([a-zA-Z0-9_]{3,30})`)

type CommentUseCase struct {
	uow          domain.UnitOfWorkInterface
	cursorSecret []byte
}

func NewCommentUseCase(uow domain.UnitOfWorkInterface, cursorSecret []byte) *CommentUseCase {
	return &CommentUseCase{uow: uow, cursorSecret: cursorSecret}
}

func (uc *CommentUseCase) CreateComment(ctx context.Context, postID, authorID uuid.UUID, parentID *uuid.UUID, content string) (*entity.Comment, error) {
	content = strings.TrimSpace(content)

	if content == "" {
		return nil, commonapperr.Validation(commonapperr.CodeFieldRequired, "content", "content is required")
	}

	comment := &entity.Comment{ID: uuid.New(), PostID: postID, AuthorID: authorID, ParentID: parentID, Content: content}
	err := uc.uow.Do(ctx, func(tx domain.Tx) error {
		if err := tx.Comments().CreateComment(ctx, comment); err != nil {
			return err
		}

		if comment.ReplyToUserID != nil && *comment.ReplyToUserID != authorID {
			parent := ""
			if comment.ParentID != nil {
				parent = comment.ParentID.String()
			}

			payload, err := buildEnvelope(events.EventCommentReplied, events.CommentRepliedEvent{PostID: comment.PostID.String(), CommentID: comment.ID.String(), ParentCommentID: parent, ReplyAuthorID: authorID.String(), OriginalAuthorID: comment.ReplyToUserID.String()})

			if err != nil {
				return err
			}

			if err = tx.Outbox().Insert(ctx, &outbox.OutboxEvent{ID: uuid.New(), AggregateID: comment.ID.String(), AggregateType: "comment", EventType: events.EventCommentReplied, Payload: payload}); err != nil {
				return err
			}
		}

		mentions := extractMentions(comment.Content)

		if len(mentions) > 0 {
			payload, err := buildEnvelope(events.EventCommentMention, events.CommentMentionedEvent{
				PostID:        comment.PostID.String(),
				CommentID:     comment.ID.String(),
				AuthorID:      authorID.String(),
				MentionedList: mentions,
			})

			if err != nil {
				return err
			}

			if err = tx.Outbox().Insert(ctx, &outbox.OutboxEvent{
				ID:            uuid.New(),
				AggregateID:   comment.ID.String(),
				AggregateType: "comment",
				EventType:     events.EventCommentMention,
				Payload:       payload,
			}); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return comment, nil
}

func (uc *CommentUseCase) GetPostComments(ctx context.Context, postID uuid.UUID, limit int, cursorToken string) (domain.CommentsResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = defaultLimit
	}

	after, afterID, err := uc.decodeCommentsCursor(cursorToken)

	if err != nil {
		return domain.CommentsResponse{}, commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid cursor")
	}

	comments, err := uc.uow.CommentReader().GetCommentsByPostID(ctx, postID, limit, after, afterID)

	if err != nil {
		return domain.CommentsResponse{}, err
	}

	resp := domain.CommentsResponse{Comments: comments}

	if len(comments) == limit {
		last := comments[len(comments)-1]
		token, _ := cursor.Encode(uc.cursorSecret, last.CreatedAt, last.ID)
		resp.NextCursor = token
	}

	return resp, nil
}

func (uc *CommentUseCase) UpdateComment(ctx context.Context, commentID, authorID uuid.UUID, content string) (*entity.Comment, error) {
	content = strings.TrimSpace(content)

	if content == "" {
		return nil, commonapperr.Validation(commonapperr.CodeFieldRequired, "content", "content is required")
	}

	comment := &entity.Comment{ID: commentID, AuthorID: authorID, Content: content}

	if err := uc.uow.CommentReader().UpdateComment(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (uc *CommentUseCase) UpvoteComment(ctx context.Context, commentID, userID uuid.UUID) (*entity.Comment, error) {
	return uc.uow.CommentReader().SetCommentVote(ctx, commentID, userID, 1)
}
func (uc *CommentUseCase) DownvoteComment(ctx context.Context, commentID, userID uuid.UUID) (*entity.Comment, error) {
	return uc.uow.CommentReader().SetCommentVote(ctx, commentID, userID, -1)
}
func (uc *CommentUseCase) RemoveCommentVote(ctx context.Context, commentID, userID uuid.UUID) (*entity.Comment, error) {
	return uc.uow.CommentReader().RemoveCommentVote(ctx, commentID, userID)
}
func (uc *CommentUseCase) DeleteComment(ctx context.Context, commentID, authorID uuid.UUID) error {
	return uc.uow.CommentReader().DeleteComment(ctx, commentID, authorID)
}

func (uc *CommentUseCase) decodeCommentsCursor(token string) (time.Time, uuid.UUID, error) {
	if token == "" {
		return time.Unix(0, 0).UTC(), uuid.Nil, nil
	}

	return cursor.Decode(uc.cursorSecret, token)
}

func extractMentions(content string) []string {
	matches := mentionRegexp.FindAllStringSubmatch(content, -1)

	if len(matches) == 0 {
		return nil
	}

	uniq := make(map[string]struct{}, len(matches))
	result := make([]string, 0, len(matches))

	for _, m := range matches {
		if len(m) < 2 {
			continue
		}

		v := strings.ToLower(strings.TrimSpace(m[1]))

		if v == "" {
			continue
		}

		if _, ok := uniq[v]; ok {
			continue
		}

		uniq[v] = struct{}{}
		result = append(result, v)
	}

	return result
}
