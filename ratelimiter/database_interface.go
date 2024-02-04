package ratelimiter

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type DataLimiter interface {
	ZRemRangeByScore(ctx context.Context, key, min, max string) (int64, error)
	ZCard(ctx context.Context, key string) (int64, error)
	ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error)
	SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}
