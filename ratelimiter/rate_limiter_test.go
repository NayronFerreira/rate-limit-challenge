package ratelimiter

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/NayronFerreira/rate-limit-challenge/infra/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckRateLimitForKey_NotExceeded(t *testing.T) {
	ctx := context.Background()
	mockRedis := new(database.MockRedisClient)
	db := NewLimiter(mockRedis, map[string]int64{"test_token": 2}, 1, 5, 3)

	// Configura expectativas para chamadas ao Redis
	mockRedis.On("Exists", ctx, []string{"block:test_token"}).Return(int64(0), nil) // Simula chave não bloqueada
	mockRedis.On("ZRemRangeByScore", ctx, "limiter:test_token", "-inf", mock.Anything).Return(int64(0), nil)
	mockRedis.On("ZCard", ctx, "limiter:test_token").Return(int64(1), nil) // Simula contagem abaixo do limite
	mockRedis.On("Get", mock.AnythingOfType("*context.timerCtx"), "test_token").Return(`{"token":"test_token","limitReq":2}`, nil)
	mockRedis.On("ZAdd", ctx, "limiter:test_token", mock.Anything).Return(int64(1), nil)

	exceeded, err := db.CheckRateLimitForKey(ctx, "test_token", true)
	assert.NoError(t, err)
	assert.False(t, exceeded)

	mockRedis.AssertExpectations(t)
}

func TestBlockKey(t *testing.T) {
	mockRedis := new(database.MockRedisClient)
	db := NewLimiter(mockRedis, nil, 60, 60, 10)
	ctx := context.Background()

	mockRedis.On("SetEX", ctx, "block:test_key", "", time.Minute).Return(nil)

	err := db.BlockKey(ctx, "test_key")
	assert.NoError(t, err)
}

func TestIsKeyBlocked_Blocked(t *testing.T) {
	mockRedis := new(database.MockRedisClient)
	db := NewLimiter(mockRedis, nil, 60, 60, 10)
	ctx := context.Background()

	mockRedis.On("Exists", ctx, []string{"block:test_key"}).Return(int64(1), nil)

	blocked, err := db.IsKeyBlocked(ctx, "test_key")
	assert.NoError(t, err)
	assert.True(t, blocked)
}

func TestRegisterPersonalizedTokens(t *testing.T) {
	mockRedis := new(database.MockRedisClient)
	db := NewLimiter(mockRedis, map[string]int64{"custom_token": 5}, 1, 5, 3)
	ctx := context.Background()

	tokenData := struct {
		Token    string `json:"token"`
		LimitReq int64  `json:"limitReq"`
	}{
		Token:    "custom_token",
		LimitReq: 5,
	}
	jsonData, _ := json.Marshal(tokenData)

	mockRedis.On("Set", ctx, "custom_token", jsonData, time.Duration(0)).Return(nil)
	mockRedis.On("Get", ctx, "custom_token").Return(`{"token":"custom_token","limitReq":2}`, nil)

	err := db.RegisterPersonalizedTokens(ctx)
	assert.NoError(t, err)
}

func TestRegisterPersonalizedTokens_SetError(t *testing.T) {
	mockRedis := new(database.MockRedisClient)
	db := NewLimiter(mockRedis, map[string]int64{"custom_token": 5}, 1, 5, 3)
	ctx := context.Background()

	tokenData := struct {
		Token    string `json:"token"`
		LimitReq int64  `json:"limitReq"`
	}{
		Token:    "custom_token",
		LimitReq: 5,
	}
	jsonData, _ := json.Marshal(tokenData)

	mockRedis.On("Set", ctx, "custom_token", jsonData, time.Duration(0)).Return(errors.New("set error"))

	err := db.RegisterPersonalizedTokens(ctx)
	assert.Error(t, err)
}
