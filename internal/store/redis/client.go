package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

func NewClient(ctx context.Context, redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	opts.PoolSize = 50
	opts.MinIdleConns = 10
	opts.ConnMaxLifetime = 5 * time.Minute
	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
