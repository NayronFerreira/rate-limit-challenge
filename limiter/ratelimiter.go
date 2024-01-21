package limiter

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type Limiter struct {
	RedisClient               *redis.Client
	IPMaxRequestsPerSecond    int
	TokenMaxRequestsPerSecond int
	LockDurationInSeconds     int
	BlockDurationInSeconds    int
}

func NewLimiter(redisClient *redis.Client, tokenMaxRequestsPerSecond, ipMaxRequestsPerSecond, lockDurationSeconds, blockDurationSeconds int) *Limiter {
	return &Limiter{
		RedisClient:               redisClient,
		IPMaxRequestsPerSecond:    ipMaxRequestsPerSecond,
		TokenMaxRequestsPerSecond: tokenMaxRequestsPerSecond,
		LockDurationInSeconds:     lockDurationSeconds,
		BlockDurationInSeconds:    blockDurationSeconds,
	}
}

// CheckRateLimit verifica se uma requisição excede o limite de taxa configurado para um determinado IP ou token.
func (l *Limiter) CheckRateLimit(ctx context.Context, key string, isToken bool) (bool, error) {

	isBlocked, err := l.IsKeyBlocked(ctx, key)
	if err != nil {
		return false, err
	}

	if isBlocked {
		return true, nil
	}

	redisKey := "limiter:" + key

	now := time.Now().Unix() // Obtém o tempo agora em segundos desde a epoch
	minScore := "-inf"

	// Remova os membros do conjunto cujo score é menor que o tempo agora
	_, err = l.RedisClient.ZRemRangeByScore(ctx, redisKey, minScore, strconv.FormatInt(now, 10)).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}

	// Verifique o número de membros restantes no conjunto
	cmd := l.RedisClient.ZCard(ctx, redisKey)
	count, err := cmd.Result()
	if err != nil && err != redis.Nil {
		return false, err
	}

	var reqRateLimit int
	if isToken {
		reqRateLimit = l.TokenMaxRequestsPerSecond
	} else {
		reqRateLimit = l.IPMaxRequestsPerSecond
	}

	if count < int64(reqRateLimit) {

		log.Printf("key: %s count: %d, reqLimit: %d \n", key, count, reqRateLimit)
		expireTime := now + int64(l.LockDurationInSeconds)

		_, err := l.RedisClient.ZAdd(ctx, redisKey, &redis.Z{
			Score:  float64(expireTime),
			Member: time.Now().Format(time.RFC3339Nano),
		}).Result()
		if err != nil {
			return false, err
		}

		return false, nil
	}

	if err = l.BlockKey(ctx, key); err != nil {
		return false, err
	}
	log.Printf("key blocked: %s count: %d, reqLimit: %d \n", key, count, reqRateLimit)

	return true, nil
}

// BlockKey bloqueia uma determinada chave.
func (l *Limiter) BlockKey(ctx context.Context, key string) error {
	// Adiciona a chave ao conjunto de chaves bloqueadas com tempo de expiração.
	return l.RedisClient.SetEX(ctx, "block:"+key, "", time.Second*time.Duration(l.BlockDurationInSeconds)).Err()
}

// IsKeyBlocked verifica se uma determinada chave está bloqueada.
func (l *Limiter) IsKeyBlocked(ctx context.Context, key string) (bool, error) {
	// Verifica se a chave já está no conjunto de chaves bloqueadas.
	exists, err := l.RedisClient.Exists(ctx, "block:"+key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}
