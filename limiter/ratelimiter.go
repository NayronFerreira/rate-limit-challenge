package limiter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type Limiter struct {
	RedisClient            *redis.Client
	ConfigToken            map[string]int64
	lockDurationSeconds    int64
	blockDurationSeconds   int64
	ipMaxRequestsPerSecond int64
}

func NewLimiter(redisClient *redis.Client, configToken map[string]int64, lockDurationSeconds, blockDurationSeconds, ipMaxRequestsPerSecond int64) *Limiter {
	limiter := &Limiter{
		RedisClient:            redisClient,
		ConfigToken:            configToken,
		lockDurationSeconds:    lockDurationSeconds,
		blockDurationSeconds:   blockDurationSeconds,
		ipMaxRequestsPerSecond: ipMaxRequestsPerSecond,
	}
	return limiter
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

		ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Second)
		defer cancel()

		tokenConfigStr, err := l.RedisClient.Get(ctx, key).Result()
		if err == redis.Nil {
			return false, errors.New("token não encontrado")
		}

		type TokenConfig struct {
			Token    string `json:"token"`
			LimitReq int64  `json:"limitReq"`
		}

		var tokenConfig TokenConfig
		if err = json.Unmarshal([]byte(tokenConfigStr), &tokenConfig); err != nil {
			return false, err
		}
		reqRateLimit = int(tokenConfig.LimitReq)

	} else {
		reqRateLimit = int(l.ipMaxRequestsPerSecond)
	}

	if count < int64(reqRateLimit) {
		log.Printf("key: %s count: %d, reqLimit: %d \n", key, count, reqRateLimit)
		expireTime := now + int64(l.lockDurationSeconds)

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

func (l *Limiter) BlockKey(ctx context.Context, key string) error {
	return l.RedisClient.SetEX(ctx, "block:"+key, "", time.Second*time.Duration(l.blockDurationSeconds)).Err()
}

func (l *Limiter) IsKeyBlocked(ctx context.Context, key string) (bool, error) {
	exists, err := l.RedisClient.Exists(ctx, "block:"+key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

func (l *Limiter) RegisterToken(ctx context.Context) error {

	for token, limitReq := range l.ConfigToken {

		data := struct {
			Token    string `json:"token"`
			LimitReq int64  `json:"limitReq"`
		}{
			Token:    token,
			LimitReq: limitReq,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}

		if err = l.RedisClient.Set(ctx, token, jsonData, 0).Err(); err != nil {
			return err
		}

		storedValue, err := l.RedisClient.Get(ctx, token).Result()
		if err != nil {
			return err
		}

		fmt.Println("storedValue: ", storedValue)
	}
	return nil
}
