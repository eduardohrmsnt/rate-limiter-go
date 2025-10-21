package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eduardohermesneto/rate-limiter/internal/infra/storage"
	"github.com/eduardohermesneto/rate-limiter/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware_IPLimit(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := usecase.NewRateLimiter(store, 3, 10, 5*time.Second)
	middleware := NewRateLimiterMiddleware(limiter)

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Contains(t, rec.Body.String(), Message429)
}

func TestMiddleware_TokenLimit(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := usecase.NewRateLimiter(store, 3, 5, 5*time.Second)
	middleware := NewRateLimiterMiddleware(limiter)

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	token := "test-token-123"

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(HeaderAPIKey, token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(HeaderAPIKey, token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Contains(t, rec.Body.String(), Message429)
}

func TestMiddleware_TokenOverridesIP(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	limiter := usecase.NewRateLimiter(store, 2, 10, 5*time.Second)
	middleware := NewRateLimiterMiddleware(limiter)

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		req.Header.Set(HeaderAPIKey, "premium-token")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code, "Request %d should succeed with token", i+1)
	}
}

func TestExtractIP(t *testing.T) {
	tests := []struct {
		name          string
		remoteAddr    string
		xForwardedFor string
		xRealIP       string
		expectedIP    string
	}{
		{
			name:       "RemoteAddr only",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:          "X-Forwarded-For single",
			remoteAddr:    "192.168.1.1:12345",
			xForwardedFor: "10.0.0.1",
			expectedIP:    "10.0.0.1",
		},
		{
			name:          "X-Forwarded-For multiple",
			remoteAddr:    "192.168.1.1:12345",
			xForwardedFor: "10.0.0.1, 10.0.0.2, 10.0.0.3",
			expectedIP:    "10.0.0.1",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "192.168.1.1:12345",
			xRealIP:    "10.0.0.5",
			expectedIP: "10.0.0.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := extractIP(req)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}
