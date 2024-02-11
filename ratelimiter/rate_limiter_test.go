package ratelimiter

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/NayronFerreira/rate-limit-challenge/infra/database"
	"github.com/go-redis/redis/v8"
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

func TestCheckRateLimitForKey_Exceeded(t *testing.T) {
	ctx := context.Background()
	mockRedis := new(database.MockRedisClient)
	db := NewLimiter(mockRedis, map[string]int64{"test_token": 2}, 1, 5, 3)

	// Configura expectativas para chamadas ao Redis corretamente
	mockRedis.On("Exists", ctx, mock.AnythingOfType("[]string")).Return(int64(0), nil)
	mockRedis.On("ZRemRangeByScore", ctx, "limiter:test_token", "-inf", mock.AnythingOfType("string")).Return(int64(0), nil)
	mockRedis.On("ZCard", ctx, "limiter:test_token").Return(int64(3), nil)
	mockRedis.On("Get", mock.AnythingOfType("*context.timerCtx"), "test_token").Return(`{"token":"test_token","limitReq":2}`, nil)
	// mockRedis.On("ZAdd", mock.Anything, "limiter:test_token", mock.AnythingOfType("[]*redis.Z")).Return(int64(1), nil)

	mockRedis.On("SetEX", ctx, "block:test_token", "", time.Duration(5)*time.Second).Return(nil)

	exceeded, err := db.CheckRateLimitForKey(ctx, "test_token", true)
	assert.NoError(t, err)
	assert.True(t, exceeded)

	mockRedis.AssertExpectations(t)
}

func TestIsRateLimitExceeded(t *testing.T) {
	ctx := context.Background()
	mockRedis := new(database.MockRedisClient)
	db := NewLimiter(mockRedis, map[string]int64{"test_token": 2}, 1, 5, 3)

	// Configura expectativas para chamadas ao Redis
	mockRedis.On("Exists", ctx, []string{"block:test_token"}).Return(int64(0), nil) // Adiciona expectativa para Exists
	mockRedis.On("ZRemRangeByScore", ctx, "limiter:test_token", "-inf", mock.Anything).Return(int64(0), nil)
	mockRedis.On("ZCard", ctx, "limiter:test_token").Return(int64(2), nil) // Ajusta a contagem para ser igual ao limite de taxa
	mockRedis.On("Get", mock.AnythingOfType("*context.timerCtx"), "test_token").Return(`{"token":"test_token","limitReq":2}`, nil)
	mockRedis.On("SetEX", ctx, "block:test_token", "", time.Duration(5*time.Second)).Return(nil) // Adiciona expectativa para SetEX

	exceeded, err := db.IsRateLimitExceeded(ctx, "test_token", true)
	assert.NoError(t, err)
	assert.True(t, exceeded) // Ajusta a expectativa para ser true, já que o limite de taxa é excedido

	mockRedis.AssertExpectations(t)
}

func TestIsRateLimitExceeded_WhenIsKeyBlockedReturnsError(t *testing.T) {
	ctx := context.Background()
	mockRedis := new(database.MockRedisClient)
	db := NewLimiter(mockRedis, map[string]int64{"test_token": 2}, 1, 5, 3)

	mockRedis.On("Exists", ctx, []string{"block:test_token"}).Return(int64(0), errors.New("mock error"))
	_, err := db.IsRateLimitExceeded(ctx, "test_token", true)
	assert.Error(t, err, "Expected error when IsKeyBlocked returns an error")
	mockRedis.AssertExpectations(t)
}

func TestIsRateLimitExceeded_WhenZRemRangeByScoreReturnsError(t *testing.T) {
	ctx := context.Background()
	mockRedis := new(database.MockRedisClient)
	db := NewLimiter(mockRedis, map[string]int64{"test_token": 2}, 1, 5, 3)

	mockRedis.On("Exists", ctx, []string{"block:test_token"}).Return(int64(0), nil)
	mockRedis.On("ZRemRangeByScore", ctx, "limiter:test_token", "-inf", mock.Anything).Return(int64(0), errors.New("mock error"))
	_, err := db.IsRateLimitExceeded(ctx, "test_token", true)
	assert.Error(t, err, "Expected error when ZRemRangeByScore returns an error")
	mockRedis.AssertExpectations(t)
}

func TestIsRateLimitExceeded_WhenZCardReturnsError(t *testing.T) {
	ctx := context.Background()
	mockRedis := new(database.MockRedisClient)
	db := NewLimiter(mockRedis, map[string]int64{"test_token": 2}, 1, 5, 3)

	mockRedis.On("Exists", ctx, []string{"block:test_token"}).Return(int64(0), nil)
	mockRedis.On("ZRemRangeByScore", ctx, "limiter:test_token", "-inf", mock.Anything).Return(int64(0), nil)
	mockRedis.On("ZCard", ctx, "limiter:test_token").Return(int64(0), errors.New("mock error"))
	_, err := db.IsRateLimitExceeded(ctx, "test_token", true)
	assert.Error(t, err, "Expected error when ZCard returns an error")
	mockRedis.AssertExpectations(t)
}

// func TestIsRateLimitExceeded_WhenGetReturnsError(t *testing.T) {
// 	ctx := context.Background()
// 	mockRedis := new(database.MockRedisClient)
// 	db := NewLimiter(mockRedis, map[string]int64{"test_token": 2}, 1, 5, 3)

// 	mockRedis.On("ZRemRangeByScore", ctx, "limiter:test_token", "-inf", mock.Anything).Return(int64(0), nil)
// 	mockRedis.On("Exists", ctx, []string{"block:test_token"}).Return(int64(0), nil)
// 	mockRedis.On("ZCard", ctx, "limiter:test_token").Return(int64(0), nil)
// 	mockRedis.On("ZAdd", ctx, "limiter:test_token", mock.Anything).Return(int64(0), nil)
// 	mockRedis.On("Get", mock.AnythingOfType("*context.timerCtx"), "test_token").Return("", errors.New("mock error")).Once()
// 	_, err := db.IsRateLimitExceeded(ctx, "test_token", false)
// 	assert.Error(t, err, "Expected error when Get returns an error")
// 	mockRedis.AssertExpectations(t)
// }

func TestIsRateLimitExceeded_WhenGetReturnsRedisNil(t *testing.T) {
	ctx := context.Background()
	mockRedis := new(database.MockRedisClient)
	db := NewLimiter(mockRedis, map[string]int64{"test_token": 2}, 1, 5, 3)

	mockRedis.On("ZCard", ctx, "limiter:test_token").Return(int64(2), nil).Once()
	mockRedis.On("Exists", ctx, []string{"block:test_token"}).Return(int64(0), nil)
	mockRedis.On("ZRemRangeByScore", ctx, "limiter:test_token", "-inf", mock.Anything).Return(int64(0), nil)
	mockRedis.On("Get", mock.AnythingOfType("*context.timerCtx"), "test_token").Return("", redis.Nil)
	_, err := db.IsRateLimitExceeded(ctx, "test_token", true)
	assert.Error(t, err, "Expected error when Get returns redis.Nil")
	mockRedis.AssertExpectations(t)
}
func TestIsRateLimitExceeded_WhenJsonUnmarshalReturnsError(t *testing.T) {
	ctx := context.Background()
	mockRedis := new(database.MockRedisClient)
	db := NewLimiter(mockRedis, map[string]int64{"test_token": 2}, 1, 5, 3)

	mockRedis.On("Exists", ctx, []string{"block:test_token"}).Return(int64(0), nil)
	mockRedis.On("ZRemRangeByScore", ctx, "limiter:test_token", "-inf", mock.Anything).Return(int64(0), nil)
	mockRedis.On("ZCard", ctx, "limiter:test_token").Return(int64(0), nil)
	mockRedis.On("Get", mock.AnythingOfType("*context.timerCtx"), "test_token").Return("invalid json", nil)
	_, err := db.IsRateLimitExceeded(ctx, "test_token", true)
	assert.Error(t, err, "Expected error when json.Unmarshal returns an error")
	mockRedis.AssertExpectations(t)
}

func TestIsRateLimitExceeded_WhenBlockKeyReturnsError(t *testing.T) {
	ctx := context.Background()
	mockRedis := new(database.MockRedisClient)
	db := NewLimiter(mockRedis, map[string]int64{"test_token": 2}, 1, 5, 3)

	// Add these lines to mock the Exists, ZRemRangeByScore, ZCard and ZAdd methods
	mockRedis.On("Exists", ctx, []string{"block:test_token"}).Return(int64(0), nil)
	mockRedis.On("ZRemRangeByScore", ctx, "limiter:test_token", "-inf", mock.Anything).Return(int64(0), nil)
	mockRedis.On("ZCard", ctx, "limiter:test_token").Return(int64(2), nil) // Return 2 for ZCard

	mockRedis.On("Get", mock.AnythingOfType("*context.timerCtx"), "test_token").Return(`{"token":"test_token","limitReq":2}`, nil)
	mockRedis.On("SetEX", ctx, "block:test_token", "", time.Duration(5*time.Second)).Return(errors.New("mock error"))

	// Call IsRateLimitExceeded twice to make a new request
	db.IsRateLimitExceeded(ctx, "test_token", true)
	_, err := db.IsRateLimitExceeded(ctx, "test_token", true)
	assert.Error(t, err, "Expected error when BlockKey returns an error")
	mockRedis.AssertExpectations(t)
}

func TestCheckRateLimitForKey_Error(t *testing.T) {
	ctx := context.Background()
	mockRedis := new(database.MockRedisClient)
	db := NewLimiter(mockRedis, map[string]int64{"test_token": 2}, 1, 5, 3)

	// Configura expectativas para chamadas ao Redis
	mockRedis.On("Exists", ctx, []string{"block:test_token"}).Return(int64(0), nil) // Simula chave não bloqueada
	mockRedis.On("ZRemRangeByScore", ctx, "limiter:test_token", "-inf", mock.Anything).Return(int64(0), errors.New("redis error"))

	exceeded, err := db.CheckRateLimitForKey(ctx, "test_token", true)
	assert.Error(t, err)
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

func TestTokenExists(t *testing.T) {
	mockRedis := new(database.MockRedisClient)
	db := NewLimiter(mockRedis, map[string]int64{"existing_token": 5}, 1, 5, 3)

	assert.True(t, db.TokenExists("existing_token"), "Expected true for existing token")
	assert.False(t, db.TokenExists("non_existing_token"), "Expected false for non-existing token")
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
