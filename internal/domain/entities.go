package domain

import "time"

type RateLimitType string

const (
	RateLimitTypeIP    RateLimitType = "ip"
	RateLimitTypeToken RateLimitType = "token"
)

type RateLimitConfig struct {
	Key           string
	Type          RateLimitType
	MaxRequests   int
	BlockDuration time.Duration
}

type RateLimitStatus struct {
	Allowed       bool
	RemainingReqs int
	BlockedUntil  time.Time
}
