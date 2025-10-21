package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/eduardohermesneto/rate-limiter/internal/domain"
)

type RateLimiter struct {
	storage       domain.Storage
	ipLimit       int
	tokenLimit    int
	blockDuration time.Duration
	tokenLimits   map[string]int
}

func NewRateLimiter(storage domain.Storage, ipLimit, tokenLimit int, blockDuration time.Duration) *RateLimiter {
	return &RateLimiter{
		storage:       storage,
		ipLimit:       ipLimit,
		tokenLimit:    tokenLimit,
		blockDuration: blockDuration,
		tokenLimits:   make(map[string]int),
	}
}

func (rl *RateLimiter) SetTokenLimit(token string, limit int) {
	rl.tokenLimits[token] = limit
}

func (rl *RateLimiter) CheckLimit(ctx context.Context, config domain.RateLimitConfig) (*domain.RateLimitStatus, error) {
	blockKey := fmt.Sprintf("block:%s:%s", config.Type, config.Key)

	blocked, err := rl.storage.IsBlocked(ctx, blockKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check block status: %w", err)
	}

	if blocked {
		ttl, err := rl.storage.GetTTL(ctx, blockKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get TTL: %w", err)
		}

		return &domain.RateLimitStatus{
			Allowed:       false,
			RemainingReqs: 0,
			BlockedUntil:  time.Now().Add(ttl),
		}, nil
	}

	countKey := fmt.Sprintf("count:%s:%s", config.Type, config.Key)

	count, err := rl.storage.Increment(ctx, countKey, time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to increment counter: %w", err)
	}

	if int(count) > config.MaxRequests {
		if err := rl.storage.SetBlock(ctx, blockKey, config.BlockDuration); err != nil {
			return nil, fmt.Errorf("failed to set block: %w", err)
		}

		return &domain.RateLimitStatus{
			Allowed:       false,
			RemainingReqs: 0,
			BlockedUntil:  time.Now().Add(config.BlockDuration),
		}, nil
	}

	remaining := config.MaxRequests - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return &domain.RateLimitStatus{
		Allowed:       true,
		RemainingReqs: remaining,
		BlockedUntil:  time.Time{},
	}, nil
}

func (rl *RateLimiter) CheckIP(ctx context.Context, ip string) (*domain.RateLimitStatus, error) {
	config := domain.RateLimitConfig{
		Key:           ip,
		Type:          domain.RateLimitTypeIP,
		MaxRequests:   rl.ipLimit,
		BlockDuration: rl.blockDuration,
	}
	return rl.CheckLimit(ctx, config)
}

func (rl *RateLimiter) CheckToken(ctx context.Context, token string) (*domain.RateLimitStatus, error) {
	limit := rl.tokenLimit
	if customLimit, exists := rl.tokenLimits[token]; exists {
		limit = customLimit
	}

	config := domain.RateLimitConfig{
		Key:           token,
		Type:          domain.RateLimitTypeToken,
		MaxRequests:   limit,
		BlockDuration: rl.blockDuration,
	}
	return rl.CheckLimit(ctx, config)
}
