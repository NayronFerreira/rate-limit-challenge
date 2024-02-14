package contract_db

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Datastore interface {
	// This ZRemRangeByScore method is used to remove all members in a sorted set within the given scores.
	ZRemRangeByScore(ctx context.Context, key, min, max string) (int64, error)

	// This ZCard method is used to get the number of members in a sorted set.
	ZCard(ctx context.Context, key string) (int64, error)

	// This ZAdd method is used to add one or more members to a sorted set, or update its score if it already exists.
	ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error)

	// This SetEX method is used to set the value and expiration of a key.
	SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// This Exists method is used to check if a key exists.
	Exists(ctx context.Context, keys ...string) (int64, error)

	// This Get method is used to get the value of a key.
	Get(ctx context.Context, key string) (string, error)

	// This Set method is used to set the value of a key.
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}
