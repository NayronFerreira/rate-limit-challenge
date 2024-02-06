package ratelimiter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/NayronFerreira/rate-limit-challenge/ratelimiter/contract_db"
	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	Database               interface{ contract_db.Datastore }
	ConfigToken            map[string]int64
	lockDurationSeconds    int64
	blockDurationSeconds   int64
	ipMaxRequestsPerSecond int64
}

func NewLimiter(db contract_db.Datastore, configToken map[string]int64, lockDurationSeconds, blockDurationSeconds, ipMaxRequestsPerSecond int64) *RateLimiter {
	limiter := &RateLimiter{
		Database:               db,
		ConfigToken:            configToken,
		lockDurationSeconds:    lockDurationSeconds,
		blockDurationSeconds:   blockDurationSeconds,
		ipMaxRequestsPerSecond: ipMaxRequestsPerSecond,
	}
	return limiter
}

func (l *RateLimiter) CheckRateLimitForKey(ctx context.Context, key string, isToken bool) (bool, error) {

	type result struct {
		Key     string
		Blocked bool
		Err     error
	}

	results := make(chan result, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func(key string) {
		defer wg.Done()

		isBlocked, err := l.IsRateLimitExceeded(ctx, key, isToken)

		results <- result{Key: key, Blocked: isBlocked, Err: err}
	}(key)

	go func() {
		wg.Wait()
		close(results)
	}()

	var isBlocked bool
	var err error

	for r := range results {
		if r.Err != nil {
			log.Printf("Error checking rate limit for key %s: %v", r.Key, r.Err)
			err = r.Err
		} else if r.Blocked {
			isBlocked = true
		}
	}

	return isBlocked, err
}

func (l *RateLimiter) IsRateLimitExceeded(ctx context.Context, key string, isToken bool) (bool, error) {
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
	_, err = l.Database.ZRemRangeByScore(ctx, redisKey, minScore, strconv.FormatInt(now, 10))
	if err != nil && err != redis.Nil {
		return false, err
	}

	// Verifique o número de membros restantes no conjunto
	count, err := l.Database.ZCard(ctx, redisKey)
	if err != nil && err != redis.Nil {
		return false, err
	}

	var reqRateLimit int

	if isToken {

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tokenConfigStr, err := l.Database.Get(ctx, key)
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
		log.Printf("key: %s count: %d, reqLimit: %d \n", key, count+1, reqRateLimit)
		expireTime := now + int64(l.lockDurationSeconds)

		_, err := l.Database.ZAdd(ctx, redisKey, &redis.Z{
			Score:  float64(expireTime),
			Member: time.Now().Format(time.RFC3339Nano),
		})
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

func (l *RateLimiter) BlockKey(ctx context.Context, key string) error {
	return l.Database.SetEX(ctx, "block:"+key, "", time.Second*time.Duration(l.blockDurationSeconds))
}

func (l *RateLimiter) IsKeyBlocked(ctx context.Context, key string) (bool, error) {
	exists, err := l.Database.Exists(ctx, "block:"+key)
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

func (r *RateLimiter) TokenExists(token string) bool {
	_, exists := r.ConfigToken[token]
	return exists
}

func (l *RateLimiter) RegisterPersonalizedTokens(ctx context.Context) error {

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

		if err = l.Database.Set(ctx, token, jsonData, 0); err != nil {
			return err
		}

		storedValue, err := l.Database.Get(ctx, token)
		if err != nil {
			return err
		}

		fmt.Println("storedValue: ", storedValue)
	}
	return nil
}
