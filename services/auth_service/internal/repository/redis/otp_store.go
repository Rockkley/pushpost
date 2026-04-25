package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	otpPrefix      = "otp:"
	attemptsPrefix = "otp_attempts:"
	cooldownPrefix = "otp_cooldown:"
)

var ErrOTPNotFound = errors.New("OTP not found or expired")

type OTPStore struct {
	rdb     *redis.Client
	timeout time.Duration
}

func NewOTPStore(rdb *redis.Client, timeout time.Duration) *OTPStore {
	return &OTPStore{rdb: rdb, timeout: timeout}
}

func (s *OTPStore) Save(ctx context.Context, email, code string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.rdb.Set(ctx, otpKey(email), code, ttl).Err()
}

func (s *OTPStore) Get(ctx context.Context, email string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	code, err := s.rdb.Get(ctx, otpKey(email)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrOTPNotFound
		}
		return "", fmt.Errorf("redis get otp: %w", err)
	}

	return code, nil
}

func (s *OTPStore) Delete(ctx context.Context, email string) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.rdb.Del(ctx, otpKey(email), attemptsKey(email)).Err()
}

func (s *OTPStore) IncrAttempts(ctx context.Context, email string, ttl time.Duration) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	aKey := attemptsKey(email)
	count, err := s.rdb.Incr(ctx, aKey).Result()

	if err != nil {
		return 0, fmt.Errorf("redis incr otp attempts: %w", err)
	}
	if count == 1 {
		s.rdb.Expire(ctx, aKey, ttl)
	}

	return count, nil
}

func (s *OTPStore) SetCooldown(ctx context.Context, email string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.rdb.Set(ctx, cooldownKey(email), "1", ttl).Err()
}

func (s *OTPStore) HasCooldown(ctx context.Context, email string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	n, err := s.rdb.Exists(ctx, cooldownKey(email)).Result()

	if err != nil {
		return false, fmt.Errorf("redis check otp cooldown: %w", err)
	}

	return n > 0, nil
}

func otpKey(email string) string      { return otpPrefix + email }
func attemptsKey(email string) string { return attemptsPrefix + email }
func cooldownKey(email string) string { return cooldownPrefix + email }
