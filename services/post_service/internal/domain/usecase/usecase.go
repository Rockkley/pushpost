package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rockkley/pushpost/services/post_service/internal/cache"
	"github.com/rockkley/pushpost/services/post_service/internal/domain/events"
	"strings"
	"time"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	apperr "github.com/rockkley/pushpost/services/post_service/internal/apperror"
	"github.com/rockkley/pushpost/services/post_service/internal/domain"
	"github.com/rockkley/pushpost/services/post_service/internal/entity"
	"log/slog"
)

const defaultLimit = 20

type PostUseCase struct {
	uow        domain.UnitOfWorkInterface
	friendship domain.FriendshipClient
	cache      cache.FeedCache // Redis абстракция
}

func NewPostUseCase(uow domain.UnitOfWorkInterface, friendship domain.FriendshipClient, cache cache.FeedCache) *PostUseCase {
	return &PostUseCase{uow: uow, friendship: friendship, cache: cache}
}

func (uc *PostUseCase) CreatePost(ctx context.Context, authorID uuid.UUID, content string) (*entity.Post, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "PostUseCase.CreatePost"))

	content = strings.TrimSpace(content)
	r := []rune(content)

	if len(r) == 0 {

		return nil, apperr.ContentEmpty()
	}

	if len(r) > 5000 {

		return nil, apperr.ContentTooLong()
	}

	post := &entity.Post{
		ID:       uuid.New(),
		AuthorID: authorID,
		Content:  content,
	}

	eventPayload, err := buildEventPayload("post.created", events.PostCreatedEvent{
		PostID:   post.ID.String(),
		AuthorID: post.AuthorID.String(),
	})

	if err != nil {

		return nil, err
	}

	err = uc.uow.Do(ctx, func(tx domain.Tx) error {
		if err = tx.Posts().Create(ctx, post); err != nil {

			return err
		}

		return tx.Outbox().Insert(ctx, &outbox.OutboxEvent{
			ID:            uuid.New(),
			AggregateID:   post.ID.String(),
			AggregateType: "post",
			EventType:     events.EventPostCreated,
			Payload:       eventPayload,
		})
	})

	if err != nil {
		log.Error("failed to create post", slog.Any("error", err))

		return nil, err
	}

	log.Info("post created", slog.String("post_id", post.ID.String()))

	return post, nil
}

func (uc *PostUseCase) GetFeed(ctx context.Context, userID uuid.UUID, limit int, cursorStr string) ([]*entity.Post, string, error) {

	if limit <= 0 || limit > 100 {
		limit = defaultLimit
	}

	// 1. Попытка из кеша (только первая страница — без курсора)
	if cursorStr == "" && uc.cache != nil {
		if cached, err := uc.cache.GetFeed(ctx, userID); err == nil && len(cached) > 0 {
			return cached, buildCursor(cached[len(cached)-1]), nil
		}
	}

	// 2. Получаем друзей
	friendIDs, err := uc.friendship.GetFriendIDs(ctx, userID)

	if err != nil {
		return nil, "", commonapperr.Service("failed to get friend ids", err)
	}
	// Включаем посты самого пользователя
	authorIDs := append(friendIDs, userID)

	before, beforeID, err := parseCursor(cursorStr)
	if err != nil {
		return nil, "", commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid cursor")
	}

	posts, err := uc.uow.Reader().GetByAuthors(ctx, authorIDs, limit, before, beforeID)

	if err != nil {
		return nil, "", err
	}

	if cursorStr == "" && uc.cache != nil && len(posts) > 0 {
		_ = uc.cache.SetFeed(ctx, userID, posts, 5*time.Minute)
	}

	nextCursor := ""

	if len(posts) == limit {
		nextCursor = buildCursor(posts[len(posts)-1])
	}

	return posts, nextCursor, nil
}

func (uc *PostUseCase) GetUserPosts(ctx context.Context, authorID uuid.UUID, limit int, cursorStr string) ([]*entity.Post, string, error) {
	if limit <= 0 || limit > 100 {
		limit = defaultLimit
	}

	before, beforeID, err := parseCursor(cursorStr)

	if err != nil {
		return nil, "", commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid cursor")
	}

	posts, err := uc.uow.Reader().GetByAuthor(ctx, authorID, limit, before, beforeID)

	if err != nil {
		return nil, "", err
	}

	nextCursor := ""

	if len(posts) == limit {
		nextCursor = buildCursor(posts[len(posts)-1])
	}

	return posts, nextCursor, nil
}

func (uc *PostUseCase) DeletePost(ctx context.Context, postID, authorID uuid.UUID) error {
	log := ctxlog.From(ctx).With(slog.String("op", "PostUseCase.DeletePost"))

	eventPayload, err := buildEventPayload("post.deleted", events.PostDeletedEvent{
		PostID:   postID.String(),
		AuthorID: authorID.String(),
	})

	if err != nil {
		return err
	}

	err = uc.uow.Do(ctx, func(tx domain.Tx) error {
		if err = tx.Posts().SoftDelete(ctx, postID, authorID); err != nil {
			return err
		}

		return tx.Outbox().Insert(ctx, &outbox.OutboxEvent{
			ID:            uuid.New(),
			AggregateID:   postID.String(),
			AggregateType: "post",
			EventType:     events.EventPostDeleted,
			Payload:       eventPayload,
		})
	})

	if err != nil {
		log.Error("failed to delete post", slog.Any("error", err))
		return err
	}

	log.Info("post deleted", slog.String("post_id", postID.String()))

	return nil
}

func (uc *PostUseCase) GetPostByID(ctx context.Context, postID uuid.UUID) (*entity.Post, error) {
	return uc.uow.Reader().FindByID(ctx, postID)
}

// buildCursor кодирует курсор как "timestamp|uuid"
func buildCursor(p *entity.Post) string {
	return fmt.Sprintf("%s|%s", p.CreatedAt.UTC().Format(time.RFC3339Nano), p.ID.String())
}

func parseCursor(cursor string) (before time.Time, beforeID uuid.UUID, err error) {
	if cursor == "" {
		return time.Now().Add(24 * time.Hour), uuid.Max, nil
	}

	parts := strings.SplitN(cursor, "|", 2)

	if len(parts) != 2 {
		return before, beforeID, fmt.Errorf("invalid cursor format")
	}

	before, err = time.Parse(time.RFC3339Nano, parts[0])

	if err != nil {
		return
	}

	beforeID, err = uuid.Parse(parts[1])

	return
}

func buildEventPayload(eventType string, payload any) ([]byte, error) {
	inner, err := json.Marshal(payload)

	if err != nil {
		return nil, commonapperr.Internal("marshal event inner payload", err)
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
