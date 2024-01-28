package limiter

import (
	"context"
	"log"
)

type RateLimiterStore interface {
	IncrementRequestCount(ctx context.Context, key string, isToken bool) error
	IsRateLimitExceeded(ctx context.Context, key string, isToken bool) (bool, error)
	BlockKey(ctx context.Context, key string) error
	IsKeyBlocked(ctx context.Context, key string) (bool, error)
	SetTokenRateLimit(ctx context.Context, token string, maxRequestsPerSecond int, lockDurationSeconds int) error
	GetTokenRateLimit(ctx context.Context, token string) (int, int, error)
}

type Limiter struct {
	Store                     RateLimiterStore
	IPMaxRequestsPerSecond    int
	TokenMaxRequestsPerSecond int
	LockDurationInSeconds     int
	BlockDurationInSeconds    int
}

func NewLimiter(store RateLimiterStore, tokenMaxRequestsPerSecond, ipMaxRequestsPerSecond, lockDurationSeconds, blockDurationSeconds int) *Limiter {
	return &Limiter{
		Store:                     store,
		IPMaxRequestsPerSecond:    ipMaxRequestsPerSecond,
		TokenMaxRequestsPerSecond: tokenMaxRequestsPerSecond,
		LockDurationInSeconds:     lockDurationSeconds,
		BlockDurationInSeconds:    blockDurationSeconds,
	}
}

func (l *Limiter) CheckRateLimit(ctx context.Context, key string, isToken bool) (bool, error) {
	isBlocked, err := l.Store.IsKeyBlocked(ctx, key)
	if err != nil {
		return false, err
	}

	if isBlocked {
		return true, nil
	}

	err = l.Store.IncrementRequestCount(ctx, key, isToken)
	if err != nil {
		return false, err
	}

	isExceeded, err := l.Store.IsRateLimitExceeded(ctx, key, isToken)
	if err != nil {
		return false, err
	}

	if isExceeded {
		err = l.Store.BlockKey(ctx, key)
		if err != nil {
			return false, err
		}
		log.Printf("key blocked: %s \n", key)
		return true, nil
	}

	return false, nil
}
