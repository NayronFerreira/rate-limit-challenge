package limiter

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisStore struct {
	Client                    *redis.Client
	IPMaxRequestsPerSecond    int
	TokenMaxRequestsPerSecond int
	LockDurationInSeconds     int
	BlockDurationInSeconds    int
}

func (s *RedisStore) IncrementRequestCount(ctx context.Context, key string, isToken bool) error {
	now := time.Now().Unix()
	expireTime := now + int64(s.LockDurationInSeconds)
	redisKey := "limiter:" + key

	_, err := s.Client.ZAdd(ctx, redisKey, &redis.Z{
		Score:  float64(expireTime),
		Member: time.Now().Format(time.RFC3339Nano),
	}).Result()

	return err
}

func (s *RedisStore) IsRateLimitExceeded(ctx context.Context, key string, isToken bool) (bool, error) {
	now := time.Now().Unix()
	minScore := "-inf"
	redisKey := "limiter:" + key

	_, err := s.Client.ZRemRangeByScore(ctx, redisKey, minScore, strconv.FormatInt(now, 10)).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}

	cmd := s.Client.ZCard(ctx, redisKey)
	count, err := cmd.Result()
	if err != nil && err != redis.Nil {
		return false, err
	}

	var reqRateLimit int
	if isToken {
		reqRateLimit = s.TokenMaxRequestsPerSecond
	} else {
		reqRateLimit = s.IPMaxRequestsPerSecond
	}

	return count >= int64(reqRateLimit), nil
}

func (s *RedisStore) BlockKey(ctx context.Context, key string) error {
	return s.Client.SetEX(ctx, "block:"+key, "", time.Second*time.Duration(s.BlockDurationInSeconds)).Err()
}

func (s *RedisStore) IsKeyBlocked(ctx context.Context, key string) (bool, error) {
	exists, err := s.Client.Exists(ctx, "block:"+key).Result()
	if err != nil {
		return false, err
	}

	return exists == 1, nil
}

func (s *RedisStore) SetTokenRateLimit(ctx context.Context, token string, maxRequestsPerSecond int, lockDurationSeconds int) error {
	// Chave no Redis para armazenar os dados do token
	tokenKey := "token:" + token

	// Armazenar o limite de taxa e a duração de bloqueio do token no Redis
	pipe := s.Client.TxPipeline()
	pipe.HSet(ctx, tokenKey, "maxRequestsPerSecond", maxRequestsPerSecond)
	pipe.HSet(ctx, tokenKey, "lockDurationSeconds", lockDurationSeconds)
	_, err := pipe.Exec(ctx)

	return err
}

func (s *RedisStore) GetTokenRateLimit(ctx context.Context, token string) (int, int, error) {
	// Chave no Redis que armazena os dados do token
	tokenKey := "token:" + token

	// Obter o limite de taxa e a duração de bloqueio do token do Redis
	result, err := s.Client.HMGet(ctx, tokenKey, "maxRequestsPerSecond", "lockDurationSeconds").Result()
	if err != nil {
		return 0, 0, err
	}

	maxRequestsPerSecond, err := strconv.Atoi(result[0].(string))
	if err != nil {
		return 0, 0, err
	}

	lockDurationSeconds, err := strconv.Atoi(result[1].(string))
	if err != nil {
		return 0, 0, err
	}

	return maxRequestsPerSecond, lockDurationSeconds, nil
}
