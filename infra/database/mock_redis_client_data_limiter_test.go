package database

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestZRemRangeByScoreMock(t *testing.T) {
	mockClient := new(MockRedisClient)
	ctx := context.Background()
	key := "test_key"
	min := "-inf"
	max := "100"

	mockClient.On("ZRemRangeByScore", ctx, key, min, max).Return(int64(2), nil)

	removed, err := mockClient.ZRemRangeByScore(ctx, key, min, max)

	assert.NoError(t, err, "Erro deveria ser nil")
	assert.Equal(t, int64(2), removed, "A quantidade de elementos removidos deve ser igual a 2")

	mockClient.AssertExpectations(t)
}
func TestZCardMock(t *testing.T) {
	mockClient := new(MockRedisClient)
	mockClient.On("ZCard", mock.Anything, "key1").Return(int64(3), nil)

	count, err := mockClient.ZCard(context.Background(), "key1")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)

	mockClient.AssertExpectations(t)
}

func TestZAddMock(t *testing.T) {
	mockClient := new(MockRedisClient)

	members := []*redis.Z{{Score: 1, Member: "member1"}}
	mockClient.On("ZAdd", mock.Anything, "key1", members).Return(int64(1), nil)

	added, err := mockClient.ZAdd(context.Background(), "key1", members...)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), added)

	mockClient.AssertExpectations(t)
}

func TestSetEXMock(t *testing.T) {
	mockClient := new(MockRedisClient)
	mockClient.On("SetEX", mock.Anything, "key1", "value1", time.Minute).Return(nil)

	err := mockClient.SetEX(context.Background(), "key1", "value1", time.Minute)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestExistsMock(t *testing.T) {
	mockClient := new(MockRedisClient)
	mockClient.On("Exists", mock.Anything, []string{"key1"}).Return(int64(1), nil)

	exists, err := mockClient.Exists(context.Background(), "key1")

	assert.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	mockClient.AssertExpectations(t)
}

func TestGetMock(t *testing.T) {
	mockClient := new(MockRedisClient)
	mockClient.On("Get", mock.Anything, "key1").Return("value1", nil)

	value, err := mockClient.Get(context.Background(), "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)

	mockClient.AssertExpectations(t)
}

func TestSetMock(t *testing.T) {
	mockClient := new(MockRedisClient)
	mockClient.On("Set", mock.Anything, "key1", "value1", time.Minute).Return(nil)

	err := mockClient.Set(context.Background(), "key1", "value1", time.Minute)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}
