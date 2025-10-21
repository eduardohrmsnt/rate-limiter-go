package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(host, port, password string, db int) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisStorage{
		client: client,
	}, nil
}

func (r *RedisStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	pipe := r.client.Pipeline()

	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("failed to increment and set expiration: %w", err)
	}

	return incr.Val(), nil
}

func (r *RedisStorage) Get(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get value: %w", err)
	}
	return val, nil
}

func (r *RedisStorage) SetBlock(ctx context.Context, key string, duration time.Duration) error {
	if err := r.client.Set(ctx, key, "1", duration).Err(); err != nil {
		return fmt.Errorf("failed to set block: %w", err)
	}
	return nil
}

func (r *RedisStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check block status: %w", err)
	}
	return val == "1", nil
}

func (r *RedisStorage) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}
	return ttl, nil
}

func (r *RedisStorage) Close() error {
	return r.client.Close()
}
