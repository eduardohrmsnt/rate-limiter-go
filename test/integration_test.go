package test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eduardohermesneto/rate-limiter/internal/infra/storage"
	"github.com/eduardohermesneto/rate-limiter/internal/infra/web"
	"github.com/eduardohermesneto/rate-limiter/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_FullFlow(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := usecase.NewRateLimiter(store, 5, 10, 5*time.Second)
	middleware := web.NewRateLimiterMiddleware(limiter)

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	client := &http.Client{}

	t.Run("IP rate limiting", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			resp, err := client.Get(server.URL)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}

		resp, err := client.Get(server.URL)
		require.NoError(t, err)
		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
		resp.Body.Close()
	})
}

func TestIntegration_ConcurrentRequests(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := usecase.NewRateLimiter(store, 10, 20, 5*time.Second)
	middleware := web.NewRateLimiterMiddleware(limiter)

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	results := make(chan int, 15)

	for i := 0; i < 15; i++ {
		go func() {
			resp, err := http.Get(server.URL)
			if err != nil {
				results <- 500
				return
			}
			defer resp.Body.Close()
			results <- resp.StatusCode
		}()
	}

	okCount := 0
	blockedCount := 0

	for i := 0; i < 15; i++ {
		status := <-results
		if status == http.StatusOK {
			okCount++
		} else if status == http.StatusTooManyRequests {
			blockedCount++
		}
	}

	assert.LessOrEqual(t, okCount, 10)
	assert.Greater(t, blockedCount, 0)
}

func TestIntegration_TokenPriority(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := usecase.NewRateLimiter(store, 2, 10, 5*time.Second)
	limiter.SetTokenLimit("premium", 100)

	ctx := context.Background()
	ip := "192.168.1.100"

	for i := 0; i < 2; i++ {
		status, err := limiter.CheckIP(ctx, ip)
		require.NoError(t, err)
		assert.True(t, status.Allowed)
	}

	status, err := limiter.CheckIP(ctx, ip)
	require.NoError(t, err)
	assert.False(t, status.Allowed)

	for i := 0; i < 100; i++ {
		status, err := limiter.CheckToken(ctx, "premium")
		require.NoError(t, err)
		assert.True(t, status.Allowed, fmt.Sprintf("Token request %d should be allowed", i+1))
	}
}

func TestIntegration_BlockExpiration(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := usecase.NewRateLimiter(store, 2, 10, 1*time.Second)

	ctx := context.Background()
	ip := "192.168.1.200"

	for i := 0; i < 2; i++ {
		status, err := limiter.CheckIP(ctx, ip)
		require.NoError(t, err)
		assert.True(t, status.Allowed)
	}

	status, err := limiter.CheckIP(ctx, ip)
	require.NoError(t, err)
	assert.False(t, status.Allowed)

	time.Sleep(2 * time.Second)

	status, err = limiter.CheckIP(ctx, ip)
	require.NoError(t, err)
	assert.True(t, status.Allowed)
}
