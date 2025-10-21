package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	RateLimitIP    int
	RateLimitToken int
	BlockDuration  time.Duration
	RedisHost      string
	RedisPort      string
	RedisPassword  string
	RedisDB        int
	ServerPort     string
	TokenLimits    map[string]int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	rateLimitIP, err := getEnvAsInt("RATE_LIMIT_IP", 10)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_IP: %w", err)
	}

	rateLimitToken, err := getEnvAsInt("RATE_LIMIT_TOKEN", 100)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_TOKEN: %w", err)
	}

	blockDurationSecs, err := getEnvAsInt("BLOCK_DURATION_SECONDS", 300)
	if err != nil {
		return nil, fmt.Errorf("invalid BLOCK_DURATION_SECONDS: %w", err)
	}

	redisDB, err := getEnvAsInt("REDIS_DB", 0)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}

	return &Config{
		RateLimitIP:    rateLimitIP,
		RateLimitToken: rateLimitToken,
		BlockDuration:  time.Duration(blockDurationSecs) * time.Second,
		RedisHost:      getEnv("REDIS_HOST", "localhost"),
		RedisPort:      getEnv("REDIS_PORT", "6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        redisDB,
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		TokenLimits:    make(map[string]int),
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) (int, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(valueStr)
}
