package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/post_service/internal/domain"
	"github.com/rockkley/pushpost/services/post_service/internal/domain/events"
	"github.com/rockkley/pushpost/services/post_service/internal/realtime"
	"github.com/rockkley/pushpost/services/post_service/internal/repository"
	"github.com/segmentio/kafka-go"
)

type FeedConsumer struct {
	reader     *kafka.Reader
	friendship domain.FriendshipClient
	feedRepo   repository.FeedRepository
	postRepo   repository.PostRepositoryInterface
	notifier   realtime.Notifier
	log        *slog.Logger
}

func NewFeedConsumer(
	brokers []string,
	groupID string,
	topics []string,
	friendship domain.FriendshipClient,
	feedRepo repository.FeedRepository,
	postRepo repository.PostRepositoryInterface,
	notifier realtime.Notifier,
	log *slog.Logger,
) *FeedConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		GroupTopics:    topics,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})
	return &FeedConsumer{
		reader:     reader,
		friendship: friendship,
		feedRepo:   feedRepo,
		postRepo:   postRepo,
		notifier:   notifier,
		log:        log.With("component", "feed_consumer"),
	}
}

func (c *FeedConsumer) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return fmt.Errorf("fetch message: %w", err)
		}

		if err = c.handle(ctx, msg); err != nil {
			c.log.Error("failed to handle message",
				slog.String("topic", msg.Topic),
				slog.Int64("offset", msg.Offset),
				slog.Any("error", err),
			)
			// Не коммитим — Kafka повторит доставку (at-least-once)
			continue
		}

		if err = c.reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit message: %w", err)
		}
	}
}

func (c *FeedConsumer) Close() error {
	return c.reader.Close()
}

// ── dispatch ──────────────────────────────────────────────────────────────────

type envelope struct {
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

func (c *FeedConsumer) handle(ctx context.Context, msg kafka.Message) error {
	var env envelope
	if err := json.Unmarshal(msg.Value, &env); err != nil {
		c.log.Warn("invalid envelope, skipping",
			slog.String("topic", msg.Topic),
			slog.String("value", string(msg.Value)),
		)
		return nil // битое сообщение — пропускаем без ретрая
	}

	switch msg.Topic {
	case events.EventPostCreated:
		return c.handlePostCreated(ctx, env.Payload)
	case events.EventPostUpdated:
		return c.handlePostUpdated(ctx, env.Payload)
	case events.EventPostDeleted:
		return c.handlePostDeleted(ctx, env.Payload)
	case "friendship.created":
		return c.handleFriendshipCreated(ctx, env.Payload)
	case "friendship.deleted":
		return c.handleFriendshipDeleted(ctx, env.Payload)
	case "user.deleted":
		return c.handleUserDeleted(ctx, env.Payload)
	default:
		return nil
	}
}

// ── handlers ──────────────────────────────────────────────────────────────────

func (c *FeedConsumer) handlePostCreated(ctx context.Context, payload json.RawMessage) error {
	var p events.PostCreatedEvent
	if err := json.Unmarshal(payload, &p); err != nil {
		c.log.Warn("invalid post.created payload, skipping")
		return nil
	}

	postID, err := uuid.Parse(p.PostID)
	if err != nil {
		return nil
	}
	authorID, err := uuid.Parse(p.AuthorID)
	if err != nil {
		return nil
	}

	insertedAt, err := time.Parse(time.RFC3339Nano, p.CreatedAt)
	if err != nil {
		c.log.Warn("invalid created_at in post.created, using now", slog.String("value", p.CreatedAt))
		insertedAt = time.Now().UTC()
	}

	friendIDs, err := c.friendship.GetFriendIDs(ctx, authorID)
	if err != nil {
		return fmt.Errorf("get friend ids for author %s: %w", authorID, err)
	}

	// Явная копия — не мутируем исходный slice через append
	userIDs := make([]uuid.UUID, len(friendIDs)+1)
	copy(userIDs, friendIDs)

	if err = c.feedRepo.InsertBatch(ctx, postID, userIDs, insertedAt); err != nil {
		return fmt.Errorf("feed insert batch post=%s: %w", postID, err)
	}

	// Нотифицируем всех получателей через Redis Streams
	if err = c.notifier.Publish(ctx, userIDs, realtime.FeedEvent{
		Type:   realtime.EventPostAdded,
		PostID: p.PostID,
	}); err != nil {
		c.log.Warn("notify post_added failed", slog.Any("error", err))
		// некритично — лента корректна, только SSE не дойдёт
	}

	c.log.Info("feed populated",
		slog.String("post_id", p.PostID),
		slog.Int("recipients", len(userIDs)),
	)
	return nil
}

func (c *FeedConsumer) handlePostUpdated(ctx context.Context, payload json.RawMessage) error {
	var p events.PostUpdatedEvent
	if err := json.Unmarshal(payload, &p); err != nil {
		c.log.Warn("invalid post.updated payload, skipping")
		return nil
	}

	postID, err := uuid.Parse(p.PostID)
	if err != nil {
		return nil
	}

	// Находим всех у кого этот пост в ленте
	recipients, err := c.feedRepo.FindRecipients(ctx, postID)
	if err != nil {
		return fmt.Errorf("find recipients for post %s: %w", postID, err)
	}

	if err = c.notifier.Publish(ctx, recipients, realtime.FeedEvent{
		Type:    realtime.EventPostUpdated,
		PostID:  p.PostID,
		Version: p.Version,
	}); err != nil {
		c.log.Warn("notify post_updated failed", slog.Any("error", err))
	}

	return nil
}

func (c *FeedConsumer) handlePostDeleted(ctx context.Context, payload json.RawMessage) error {
	var p events.PostDeletedEvent
	if err := json.Unmarshal(payload, &p); err != nil {
		c.log.Warn("invalid post.deleted payload, skipping")
		return nil
	}

	postID, err := uuid.Parse(p.PostID)
	if err != nil {
		return nil
	}

	// Сначала находим получателей (до удаления из feeds)
	recipients, err := c.feedRepo.FindRecipients(ctx, postID)
	if err != nil {
		return fmt.Errorf("find recipients for deleted post %s: %w", postID, err)
	}

	if err = c.feedRepo.DeleteByPostID(ctx, postID); err != nil {
		return fmt.Errorf("delete post from feeds %s: %w", postID, err)
	}

	if err = c.notifier.Publish(ctx, recipients, realtime.FeedEvent{
		Type:   realtime.EventPostDeleted,
		PostID: p.PostID,
	}); err != nil {
		c.log.Warn("notify post_deleted failed", slog.Any("error", err))
	}

	return nil
}

func (c *FeedConsumer) handleFriendshipCreated(ctx context.Context, payload json.RawMessage) error {
	var p struct {
		FriendshipID string `json:"friendship_id"`
		User1ID      string `json:"user1_id"`
		User2ID      string `json:"user2_id"`
	}

	if err := json.Unmarshal(payload, &p); err != nil {
		c.log.Warn("invalid friendship.created payload, skipping")
		return nil
	}

	user1, err := uuid.Parse(p.User1ID)

	if err != nil {
		return nil
	}

	user2, err := uuid.Parse(p.User2ID)

	if err != nil {
		return nil
	}

	if err = c.backfillFeed(ctx, user1, user2); err != nil {
		return err
	}

	return c.backfillFeed(ctx, user2, user1)
}

func (c *FeedConsumer) handleFriendshipDeleted(ctx context.Context, payload json.RawMessage) error {
	var p struct {
		UserID   string `json:"user_id"`
		FriendID string `json:"friend_id"`
	}

	if err := json.Unmarshal(payload, &p); err != nil {
		c.log.Warn("invalid friendship.deleted payload, skipping")
		return nil
	}

	user, err := uuid.Parse(p.UserID)
	if err != nil {
		return nil
	}

	friend, err := uuid.Parse(p.FriendID)
	if err != nil {
		return nil
	}

	// Убираем посты каждого из ленты другого
	if err = c.feedRepo.DeleteByAuthor(ctx, user, friend); err != nil {
		return fmt.Errorf("delete friend posts from user feed: %w", err)
	}

	if err = c.feedRepo.DeleteByAuthor(ctx, friend, user); err != nil {
		return fmt.Errorf("delete user posts from friend feed: %w", err)
	}

	// Нотифицируем обоих — фронт удаляет посты из DOM
	if err = c.notifier.Publish(ctx, []uuid.UUID{user}, realtime.FeedEvent{
		Type:     realtime.EventFriendRemoved,
		FriendID: p.FriendID,
	}); err != nil {
		c.log.Warn("notify friend_removed (user) failed", slog.Any("error", err))
	}

	if err = c.notifier.Publish(ctx, []uuid.UUID{friend}, realtime.FeedEvent{
		Type:     realtime.EventFriendRemoved,
		FriendID: p.UserID,
	}); err != nil {
		c.log.Warn("notify friend_removed (friend) failed", slog.Any("error", err))
	}

	return nil
}

func (c *FeedConsumer) handleUserDeleted(ctx context.Context, payload json.RawMessage) error {
	var p struct {
		UserID string `json:"user_id"`
	}

	if err := json.Unmarshal(payload, &p); err != nil {
		c.log.Warn("invalid user.deleted payload, skipping")

		return nil
	}

	userID, err := uuid.Parse(p.UserID)

	if err != nil {
		return nil
	}

	// Удаляем ленту самого пользователя
	if err = c.feedRepo.DeleteUserFeed(ctx, userID); err != nil {
		return fmt.Errorf("delete user feed %s: %w", userID, err)
	}

	// Удаляем посты этого пользователя из всех чужих лент
	if err = c.feedRepo.DeleteByAuthorFromAllFeeds(ctx, userID); err != nil {
		return fmt.Errorf("delete user posts from all feeds %s: %w", userID, err)
	}

	return nil
}

func (c *FeedConsumer) backfillFeed(ctx context.Context, recipientID, authorID uuid.UUID) error {
	posts, err := c.postRepo.GetByAuthor(ctx, authorID, 50, time.Now().Add(time.Hour), uuid.Max)

	if err != nil {
		return fmt.Errorf("backfill get posts author=%s: %w", authorID, err)
	}

	for _, post := range posts {
		if err = c.feedRepo.InsertBatch(ctx, post.ID, []uuid.UUID{recipientID}, post.CreatedAt); err != nil {
			return fmt.Errorf("backfill insert post=%s: %w", post.ID, err)
		}
	}

	if len(posts) > 0 {
		if err = c.notifier.Publish(ctx, []uuid.UUID{recipientID}, realtime.FeedEvent{
			Type: realtime.EventBulkNewPosts,
		}); err != nil {
			c.log.Warn("notify bulk_new_posts (backfill) failed", slog.Any("error", err))
		}
	}

	return nil
}
