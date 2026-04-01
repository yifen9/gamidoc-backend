package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

type Client struct {
	redis *goredis.Client
}

func New(addr string) *Client {
	client := goredis.NewClient(&goredis.Options{
		Addr: addr,
	})
	return &Client{redis: client}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.redis.Ping(ctx).Err()
}

func (c *Client) Ready(ctx context.Context) error {
	return c.redis.Set(ctx, "gamidoc:ready", "ok", 0).Err()
}

func (c *Client) Close() error {
	return c.redis.Close()
}
