package redis

import (
	"context"
	"crypto/tls"

	goredis "github.com/redis/go-redis/v9"
)

type Client struct {
	redis *goredis.Client
}

func New(addr, password string, useTLS bool) *Client {
	opts := &goredis.Options{
		Addr:     addr,
		Password: password,
	}
	if useTLS {
		opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	return &Client{redis: goredis.NewClient(opts)}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.redis.Ping(ctx).Err()
}

func (c *Client) Ready(ctx context.Context) error {
	return c.redis.Set(ctx, "gamidoc:ready", "ok", 0).Err()
}

func (c *Client) Raw() *goredis.Client {
	return c.redis
}

func (c *Client) Close() error {
	return c.redis.Close()
}
