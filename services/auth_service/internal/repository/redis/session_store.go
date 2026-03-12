package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain"
	"time"
)

const keyPrefix = "session:"

var ErrSessionNotFound = "session not found"

type SessionStore struct {
	redis   *redis.Client
	timeout time.Duration
}

func NewSessionStore(redis *redis.Client, timeout time.Duration) *SessionStore {
	return &SessionStore{redis: redis, timeout: timeout}
}

func (s *SessionStore) Save(ctx context.Context, session *domain.Session) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	ttl := time.Until(time.Unix(session.Expires, 0))

	if ttl <= 0 {

		return fmt.Errorf("session already expired")
	}

	data, err := json.Marshal(session)

	if err != nil {

		return fmt.Errorf("session marshal: %w", err)
	}

	if err = s.redis.Set(ctx, key(session.SessionID), data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set session: %w", err)
	}

	return nil
}

func (s *SessionStore) Get(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	data, err := s.redis.Get(ctx, key(sessionID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("%w: %s", errors.New(ErrSessionNotFound), sessionID)
		}
		return nil, fmt.Errorf("redis get session: %w", err)
	}

	var session domain.Session
	if err = json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("session unmarshal: %w", err)
	}

	return &session, nil
}

func (s *SessionStore) Delete(ctx context.Context, sessionID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := s.redis.Del(ctx, key(sessionID)).Err(); err != nil {
		return fmt.Errorf("redis delete session: %w", err)
	}

	return nil
}

func key(sessionID uuid.UUID) string {

	return keyPrefix + sessionID.String()
}
