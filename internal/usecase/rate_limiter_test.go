package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/eduardohermesneto/rate-limiter/internal/infra/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_CheckIP(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := NewRateLimiter(store, 5, 10, 5*time.Second)

	ctx := context.Background()
	ip := "192.168.1.1"

	for i := 0; i < 5; i++ {
		status, err := limiter.CheckIP(ctx, ip)
		require.NoError(t, err)
		assert.True(t, status.Allowed)
	}

	status, err := limiter.CheckIP(ctx, ip)
	require.NoError(t, err)
	assert.False(t, status.Allowed)
	assert.Equal(t, 0, status.RemainingReqs)
}

func TestRateLimiter_CheckToken(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := NewRateLimiter(store, 5, 10, 5*time.Second)

	ctx := context.Background()
	token := "test-token-123"

	for i := 0; i < 10; i++ {
		status, err := limiter.CheckToken(ctx, token)
		require.NoError(t, err)
		assert.True(t, status.Allowed)
	}

	status, err := limiter.CheckToken(ctx, token)
	require.NoError(t, err)
	assert.False(t, status.Allowed)
}

func TestRateLimiter_CustomTokenLimit(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := NewRateLimiter(store, 5, 10, 5*time.Second)
	limiter.SetTokenLimit("premium-token", 100)

	ctx := context.Background()
	token := "premium-token"

	for i := 0; i < 100; i++ {
		status, err := limiter.CheckToken(ctx, token)
		require.NoError(t, err)
		assert.True(t, status.Allowed)
	}

	status, err := limiter.CheckToken(ctx, token)
	require.NoError(t, err)
	assert.False(t, status.Allowed)
}

func TestRateLimiter_BlockDuration(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := NewRateLimiter(store, 2, 10, 2*time.Second)

	ctx := context.Background()
	ip := "192.168.1.2"

	for i := 0; i < 2; i++ {
		status, err := limiter.CheckIP(ctx, ip)
		require.NoError(t, err)
		assert.True(t, status.Allowed)
	}

	status, err := limiter.CheckIP(ctx, ip)
	require.NoError(t, err)
	assert.False(t, status.Allowed)

	time.Sleep(3 * time.Second)

	status, err = limiter.CheckIP(ctx, ip)
	require.NoError(t, err)
	assert.True(t, status.Allowed)
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := NewRateLimiter(store, 5, 10, 5*time.Second)

	ctx := context.Background()
	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	for i := 0; i < 5; i++ {
		status, err := limiter.CheckIP(ctx, ip1)
		require.NoError(t, err)
		assert.True(t, status.Allowed)
	}

	status, err := limiter.CheckIP(ctx, ip1)
	require.NoError(t, err)
	assert.False(t, status.Allowed)

	status, err = limiter.CheckIP(ctx, ip2)
	require.NoError(t, err)
	assert.True(t, status.Allowed)
}
