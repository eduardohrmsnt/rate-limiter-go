package domain

import (
	"context"
	"time"
)

type Storage interface {
	Increment(ctx context.Context, key string, expiration time.Duration) (int64, error)
	Get(ctx context.Context, key string) (int64, error)
	SetBlock(ctx context.Context, key string, duration time.Duration) error
	IsBlocked(ctx context.Context, key string) (bool, error)
	GetTTL(ctx context.Context, key string) (time.Duration, error)
	Close() error
}
