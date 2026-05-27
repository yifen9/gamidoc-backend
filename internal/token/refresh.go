package token

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
)

type RefreshStore struct {
	redis *goredis.Client
	ttl   time.Duration
}

func NewRefreshStore(redis *goredis.Client, ttl time.Duration) *RefreshStore {
	return &RefreshStore{redis: redis, ttl: ttl}
}

func (s *RefreshStore) TTL() time.Duration {
	return s.ttl
}

// GenerateAndStore creates a new opaque refresh token, stores it in Redis keyed
// to the userID, and returns the token string.
func (s *RefreshStore) GenerateAndStore(ctx context.Context, userID string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	tokenStr := hex.EncodeToString(b)

	key := refreshKey(tokenStr)
	if err := s.redis.Set(ctx, key, userID, s.ttl).Err(); err != nil {
		return "", err
	}

	return tokenStr, nil
}

// Validate looks up the refresh token and returns the associated userID.
func (s *RefreshStore) Validate(ctx context.Context, tokenStr string) (string, error) {
	key := refreshKey(tokenStr)
	userID, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return "", ErrRefreshTokenNotFound
		}
		return "", err
	}
	return userID, nil
}

// Revoke deletes a refresh token from Redis.
func (s *RefreshStore) Revoke(ctx context.Context, tokenStr string) error {
	key := refreshKey(tokenStr)
	return s.redis.Del(ctx, key).Err()
}

func refreshKey(token string) string {
	return "refresh:" + token
}
