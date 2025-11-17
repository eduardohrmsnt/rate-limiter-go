package web

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/eduardohermesneto/rate-limiter/internal/usecase"
)

const (
	HeaderAPIKey          = "API_KEY"
	Message429            = "you have reached the maximum number of requests or actions allowed within a certain time frame"
	HeaderRateLimitRemain = "X-RateLimit-Remaining"
)

type RateLimiterMiddleware struct {
	limiter *usecase.RateLimiter
}

func NewRateLimiterMiddleware(limiter *usecase.RateLimiter) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiter: limiter,
	}
}

func (m *RateLimiterMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bypass rate limiting for health checks
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.Background()

		token := r.Header.Get(HeaderAPIKey)

		if token != "" {
			status, err := m.limiter.CheckToken(ctx, token)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if !status.Allowed {
				w.Header().Set(HeaderRateLimitRemain, "0")
				http.Error(w, Message429, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		ip := extractIP(r)
		if ip == "" {
			http.Error(w, "Cannot determine IP address", http.StatusBadRequest)
			return
		}

		status, err := m.limiter.CheckIP(ctx, ip)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !status.Allowed {
			w.Header().Set(HeaderRateLimitRemain, "0")
			http.Error(w, Message429, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func extractIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
