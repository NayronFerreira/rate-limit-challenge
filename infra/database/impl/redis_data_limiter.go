package database

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisDataLimiter struct {
	client *redis.Client
}

func NewRedisDataLimiter(client *redis.Client) *RedisDataLimiter {
	return &RedisDataLimiter{client: client}
}

func (r *RedisDataLimiter) ZRemRangeByScore(ctx context.Context, key, min, max string) (int64, error) {
	return r.client.ZRemRangeByScore(ctx, key, min, max).Result()
}

func (r *RedisDataLimiter) ZCard(ctx context.Context, key string) (int64, error) {
	return r.client.ZCard(ctx, key).Result()
}

func (r *RedisDataLimiter) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	return r.client.ZAdd(ctx, key, members...).Result()
}

func (r *RedisDataLimiter) SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.SetEX(ctx, key, value, expiration).Err()
}

func (r *RedisDataLimiter) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

func (r *RedisDataLimiter) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisDataLimiter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}
