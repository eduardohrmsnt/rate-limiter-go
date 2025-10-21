package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_Increment(t *testing.T) {
	store := NewMemoryStorage()
	defer store.Close()

	ctx := context.Background()
	key := "test:key"

	val, err := store.Increment(ctx, key, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(1), val)

	val, err = store.Increment(ctx, key, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(2), val)
}

func TestMemoryStorage_Get(t *testing.T) {
	store := NewMemoryStorage()
	defer store.Close()

	ctx := context.Background()
	key := "test:key"

	store.Increment(ctx, key, 5*time.Second)
	store.Increment(ctx, key, 5*time.Second)

	val, err := store.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(2), val)
}

func TestMemoryStorage_SetBlock(t *testing.T) {
	store := NewMemoryStorage()
	defer store.Close()

	ctx := context.Background()
	key := "block:test"

	err := store.SetBlock(ctx, key, 2*time.Second)
	require.NoError(t, err)

	blocked, err := store.IsBlocked(ctx, key)
	require.NoError(t, err)
	assert.True(t, blocked)
}

func TestMemoryStorage_IsBlocked(t *testing.T) {
	store := NewMemoryStorage()
	defer store.Close()

	ctx := context.Background()
	key := "block:test"

	blocked, err := store.IsBlocked(ctx, key)
	require.NoError(t, err)
	assert.False(t, blocked)

	store.SetBlock(ctx, key, 1*time.Second)

	blocked, err = store.IsBlocked(ctx, key)
	require.NoError(t, err)
	assert.True(t, blocked)

	time.Sleep(2 * time.Second)

	blocked, err = store.IsBlocked(ctx, key)
	require.NoError(t, err)
	assert.False(t, blocked)
}

func TestMemoryStorage_GetTTL(t *testing.T) {
	store := NewMemoryStorage()
	defer store.Close()

	ctx := context.Background()
	key := "test:ttl"

	store.SetBlock(ctx, key, 5*time.Second)

	ttl, err := store.GetTTL(ctx, key)
	require.NoError(t, err)
	assert.Greater(t, ttl, 4*time.Second)
	assert.LessOrEqual(t, ttl, 5*time.Second)
}

func TestMemoryStorage_Expiration(t *testing.T) {
	store := NewMemoryStorage()
	defer store.Close()

	ctx := context.Background()
	key := "test:expire"

	store.Increment(ctx, key, 1*time.Second)

	val, err := store.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(1), val)

	time.Sleep(2 * time.Second)

	val, err = store.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(0), val)
}
