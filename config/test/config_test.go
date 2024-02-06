package config

import (
	"os"
	"testing"

	"github.com/NayronFerreira/rate-limit-challenge/config"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Define environment variables for testing
	os.Setenv("IP_MAX_REQUESTS_PER_SECOND", "100")
	os.Setenv("TOKEN_1_MAX_REQUESTS_PER_SECOND", "200")
	os.Setenv("TOKEN_2_MAX_REQUESTS_PER_SECOND", "300")
	os.Setenv("TOKEN_3_MAX_REQUESTS_PER_SECOND", "400")
	os.Setenv("TOKEN_4_MAX_REQUESTS_PER_SECOND", "500")
	os.Setenv("TOKEN_5_MAX_REQUESTS_PER_SECOND", "600")
	os.Setenv("LOCK_DURATION_SECONDS", "10")
	os.Setenv("BLOCK_DURATION_SECONDS", "20")
	os.Setenv("APP_WEB_PORT", "8080")
	os.Setenv("REDIS_URL", "redis://localhost:6379")

	config, err := config.LoadConfig()

	assert.NoError(t, err)
	assert.Equal(t, 100, config.IPMaxRequestsPerSecond)
	assert.Equal(t, int64(200), config.TokenMaxRequestsPerSecond["TOKEN_1"])
	assert.Equal(t, int64(300), config.TokenMaxRequestsPerSecond["TOKEN_2"])
	assert.Equal(t, int64(400), config.TokenMaxRequestsPerSecond["TOKEN_3"])
	assert.Equal(t, int64(500), config.TokenMaxRequestsPerSecond["TOKEN_4"])
	assert.Equal(t, int64(600), config.TokenMaxRequestsPerSecond["TOKEN_5"])
	assert.Equal(t, 10, config.LockDurationSeconds)
	assert.Equal(t, 20, config.BlockDurationSeconds)
	assert.Equal(t, "8080", config.WebPort)
	assert.Equal(t, "redis://localhost:6379", config.RedisURL)
}
