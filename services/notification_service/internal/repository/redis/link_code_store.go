package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/rockkley/pushpost/services/notification_service/internal/repository"
)

const linkCodePrefix = "tg_link:"

var ErrCodeNotFound = errors.New("link code not found or expired")

type LinkCodeStore struct{ rdb *goredis.Client }

func NewLinkCodeStore(rdb *goredis.Client) repository.LinkCodeStore { return &LinkCodeStore{rdb: rdb} }

func (s *LinkCodeStore) Save(ctx context.Context, code string, userID uuid.UUID, ttl time.Duration) error {
	return s.rdb.Set(ctx, linkCodePrefix+code, userID.String(), ttl).Err()
}

func (s *LinkCodeStore) Pop(ctx context.Context, code string) (uuid.UUID, error) {
	val, err := s.rdb.GetDel(ctx, linkCodePrefix+code).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return uuid.Nil, ErrCodeNotFound
		}
		return uuid.Nil, fmt.Errorf("redis getdel link code: %w", err)
	}
	userID, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user id in code payload: %w", err)
	}
	return userID, nil
}
