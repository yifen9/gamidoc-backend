package token

import (
	"context"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Blacklist struct {
	redis *goredis.Client
}

func NewBlacklist(redis *goredis.Client) *Blacklist {
	return &Blacklist{redis: redis}
}

// Add blacklists a JWT ID until its natural expiry time.
func (bl *Blacklist) Add(ctx context.Context, jti string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil // already expired, nothing to blacklist
	}
	return bl.redis.Set(ctx, blacklistKey(jti), "1", ttl).Err()
}

// IsBlacklisted returns true if the jti has been revoked.
func (bl *Blacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	_, err := bl.redis.Get(ctx, blacklistKey(jti)).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func blacklistKey(jti string) string {
	return "blacklist:" + jti
}
