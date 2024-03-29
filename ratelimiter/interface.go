package ratelimiter

import (
	"context"
	"time"
)

type RateLimiter interface {
	GetDecision(ctx context.Context, key string, burstSize, rate int) (time.Duration, int, error)
}
