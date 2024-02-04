package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	IPMaxRequestsPerSecond    int
	TokenMaxRequestsPerSecond map[string]int64
	LockDurationSeconds       int
	BlockDurationSeconds      int
	WebPort                   string
	RedisURL                  string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	config := &Config{
		IPMaxRequestsPerSecond: getEnvAsInt("IP_MAX_REQUESTS_PER_SECOND"),
		TokenMaxRequestsPerSecond: map[string]int64{
			"TOKEN_1": int64(getEnvAsInt("TOKEN_1_MAX_REQUESTS_PER_SECOND")),
			"TOKEN_2": int64(getEnvAsInt("TOKEN_2_MAX_REQUESTS_PER_SECOND")),
			"TOKEN_3": int64(getEnvAsInt("TOKEN_3_MAX_REQUESTS_PER_SECOND")),
			"TOKEN_4": int64(getEnvAsInt("TOKEN_4_MAX_REQUESTS_PER_SECOND")),
			"TOKEN_5": int64(getEnvAsInt("TOKEN_5_MAX_REQUESTS_PER_SECOND")),
		},
		LockDurationSeconds:  getEnvAsInt("LOCK_DURATION_SECONDS"),
		BlockDurationSeconds: getEnvAsInt("BLOCK_DURATION_SECONDS"),
		WebPort:              os.Getenv("APP_WEB_PORT"),
		RedisURL:             os.Getenv("REDIS_URL"),
	}

	return config, nil
}

func getEnvAsInt(name string) int {
	valueStr := os.Getenv(name)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Fatal("Error converting "+name+" to int:", err)
	}
	return value
}
