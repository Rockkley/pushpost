package usecase

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	apperr "github.com/rockkley/pushpost/services/post_service/internal/apperror"
	"github.com/rockkley/pushpost/services/post_service/internal/cursor"
	"github.com/rockkley/pushpost/services/post_service/internal/domain"
	"github.com/rockkley/pushpost/services/post_service/internal/domain/events"
	"github.com/rockkley/pushpost/services/post_service/internal/entity"
	"github.com/rockkley/pushpost/services/post_service/internal/repository"
	"log/slog"
)

const defaultLimit = 20

type PostUseCase struct {
	uow          domain.UnitOfWorkInterface
	feedRepo     repository.FeedRepository
	cursorSecret []byte
}

func NewPostUseCase(
	uow domain.UnitOfWorkInterface,
	feedRepo repository.FeedRepository,
	cursorSecret []byte,
) *PostUseCase {
	return &PostUseCase{uow: uow, feedRepo: feedRepo, cursorSecret: cursorSecret}
}

func (uc *PostUseCase) CreatePost(ctx context.Context, authorID uuid.UUID, content string) (*entity.Post, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "PostUseCase.CreatePost"))

	content = strings.TrimSpace(content)
	if len([]rune(content)) == 0 {
		return nil, apperr.ContentEmpty()
	}
	if len([]rune(content)) > 5000 {
		return nil, apperr.ContentTooLong()
	}

	post := &entity.Post{
		ID:       uuid.New(),
		AuthorID: authorID,
		Content:  content,
	}

	err := uc.uow.Do(ctx, func(tx domain.Tx) error {
		if err := tx.Posts().Create(ctx, post); err != nil {
			return err
		}
		// post.CreatedAt проставлен через RETURNING в репозитории
		payload, err := buildEnvelope(events.EventPostCreated, events.PostCreatedEvent{
			PostID:    post.ID.String(),
			AuthorID:  post.AuthorID.String(),
			CreatedAt: post.CreatedAt.UTC().Format(time.RFC3339Nano),
		})
		if err != nil {
			return err
		}
		return tx.Outbox().Insert(ctx, &outbox.OutboxEvent{
			ID:            uuid.New(),
			AggregateID:   post.ID.String(),
			AggregateType: "post",
			EventType:     events.EventPostCreated,
			Payload:       payload,
		})
	})
	if err != nil {
		log.Error("failed to create post", slog.Any("error", err))
		return nil, err
	}

	log.Info("post created", slog.String("post_id", post.ID.String()))
	return post, nil
}

func (uc *PostUseCase) UpdatePost(ctx context.Context, postID, authorID uuid.UUID, content string) (*entity.Post, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "PostUseCase.UpdatePost"))

	content = strings.TrimSpace(content)
	if len([]rune(content)) == 0 {
		return nil, apperr.ContentEmpty()
	}
	if len([]rune(content)) > 5000 {
		return nil, apperr.ContentTooLong()
	}

	post := &entity.Post{ID: postID, AuthorID: authorID, Content: content}

	err := uc.uow.Do(ctx, func(tx domain.Tx) error {
		if err := tx.Posts().Update(ctx, post); err != nil {
			return err
		}
		payload, err := buildEnvelope(events.EventPostUpdated, events.PostUpdatedEvent{
			PostID:  post.ID.String(),
			Version: post.Version,
		})
		if err != nil {
			return err
		}
		return tx.Outbox().Insert(ctx, &outbox.OutboxEvent{
			ID:            uuid.New(),
			AggregateID:   post.ID.String(),
			AggregateType: "post",
			EventType:     events.EventPostUpdated,
			Payload:       payload,
		})
	})
	if err != nil {
		return nil, err
	}

	log.Info("post updated", slog.String("post_id", post.ID.String()), slog.Int("version", post.Version))
	return post, nil
}

func (uc *PostUseCase) GetFeed(ctx context.Context, userID uuid.UUID, limit int, cursorToken string) (domain.FeedResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = defaultLimit
	}
	before, beforeID, err := uc.decodeCursor(cursorToken)
	if err != nil {
		return domain.FeedResponse{}, commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid cursor")
	}
	posts, err := uc.feedRepo.GetFeed(ctx, userID, limit, before, beforeID)
	if err != nil {
		return domain.FeedResponse{}, err
	}
	return uc.buildFeedResponse(posts, limit), nil
}

func (uc *PostUseCase) GetFeedSince(ctx context.Context, userID uuid.UUID, limit int, sinceToken string) (domain.FeedResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = defaultLimit
	}
	after, afterID, err := uc.decodeCursor(sinceToken)
	if err != nil {
		return domain.FeedResponse{}, commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid since cursor")
	}
	posts, err := uc.feedRepo.GetFeedSince(ctx, userID, limit, after, afterID)
	if err != nil {
		return domain.FeedResponse{}, err
	}
	return uc.buildFeedResponse(posts, limit), nil
}

func (uc *PostUseCase) GetUserPosts(ctx context.Context, authorID uuid.UUID, limit int, cursorToken string) (domain.FeedResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = defaultLimit
	}
	before, beforeID, err := uc.decodeCursor(cursorToken)
	if err != nil {
		return domain.FeedResponse{}, commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid cursor")
	}
	posts, err := uc.uow.Reader().GetByAuthor(ctx, authorID, limit, before, beforeID)
	if err != nil {
		return domain.FeedResponse{}, err
	}
	// GetByAuthor возвращает посты без InsertedAt - используем CreatedAt для курсора
	for _, p := range posts {
		p.InsertedAt = p.CreatedAt
	}
	return uc.buildFeedResponse(posts, limit), nil
}

func (uc *PostUseCase) DeletePost(ctx context.Context, postID, authorID uuid.UUID) error {
	log := ctxlog.From(ctx).With(slog.String("op", "PostUseCase.DeletePost"))

	payload, err := buildEnvelope(events.EventPostDeleted, events.PostDeletedEvent{
		PostID:   postID.String(),
		AuthorID: authorID.String(),
	})
	if err != nil {
		return err
	}

	err = uc.uow.Do(ctx, func(tx domain.Tx) error {
		if err := tx.Posts().SoftDelete(ctx, postID, authorID); err != nil {
			return err
		}
		return tx.Outbox().Insert(ctx, &outbox.OutboxEvent{
			ID:            uuid.New(),
			AggregateID:   postID.String(),
			AggregateType: "post",
			EventType:     events.EventPostDeleted,
			Payload:       payload,
		})
	})
	if err != nil {
		return err
	}

	log.Info("post deleted", slog.String("post_id", postID.String()))
	return nil
}

func (uc *PostUseCase) GetPostByID(ctx context.Context, postID uuid.UUID) (*entity.Post, error) {
	return uc.uow.Reader().FindByID(ctx, postID)
}

func (uc *PostUseCase) GetPostsByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Post, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	if len(ids) > 100 {
		return nil, commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "too many ids (max 100)")
	}
	return uc.uow.Reader().GetByIDs(ctx, ids)
}

func (uc *PostUseCase) LikePost(ctx context.Context, postID, userID uuid.UUID) (*entity.Post, error) {
	return uc.uow.Reader().SetVote(ctx, postID, userID, 1)
}

func (uc *PostUseCase) DislikePost(ctx context.Context, postID, userID uuid.UUID) (*entity.Post, error) {
	return uc.uow.Reader().SetVote(ctx, postID, userID, -1)
}

func (uc *PostUseCase) RemovePostVote(ctx context.Context, postID, userID uuid.UUID) (*entity.Post, error) {
	return uc.uow.Reader().RemoveVote(ctx, postID, userID)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (uc *PostUseCase) decodeCursor(token string) (time.Time, uuid.UUID, error) {
	if token == "" {
		ts, id := cursor.Sentinel()
		return ts, id, nil
	}
	return cursor.Decode(uc.cursorSecret, token)
}

func (uc *PostUseCase) encodeCursor(ts time.Time, id uuid.UUID) string {
	token, err := cursor.Encode(uc.cursorSecret, ts, id)
	if err != nil {
		return ""
	}
	return token
}

// buildFeedResponse строит ответ с курсорами.
// Курсор основан на InsertedAt - времени вставки в ленту, а не создания поста.
// Это обеспечивает стабильный порядок: даже если друг загрузил старый пост,
// он появится в ленте в хронологии добавления в feeds.
func (uc *PostUseCase) buildFeedResponse(posts []*entity.Post, limit int) domain.FeedResponse {
	if len(posts) == 0 {
		return domain.FeedResponse{}
	}

	resp := domain.FeedResponse{Posts: posts}

	// top_cursor - самый новый пост (первый в результате DESC)
	// Используется для GetFeedSince / reconciliation
	resp.TopCursor = uc.encodeCursor(posts[0].InsertedAt, posts[0].ID)

	// next_cursor - только если получили ровно limit записей (есть ещё страницы)
	if len(posts) == limit {
		last := posts[len(posts)-1]
		resp.NextCursor = uc.encodeCursor(last.InsertedAt, last.ID)
	}

	return resp
}

func buildEnvelope(eventType string, payload any) ([]byte, error) {
	inner, err := json.Marshal(payload)
	if err != nil {
		return nil, commonapperr.Internal("marshal event payload", err)
	}
	type envelope struct {
		EventType string          `json:"event_type"`
		Payload   json.RawMessage `json:"payload"`
	}
	result, err := json.Marshal(envelope{EventType: eventType, Payload: inner})
	if err != nil {
		return nil, commonapperr.Internal("marshal event envelope", err)
	}
	return result, nil
}
