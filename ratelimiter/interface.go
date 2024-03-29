package ratelimiter

import (
	"context"
	"time"
)

type RateLimiter interface {
	GetDecision(ctx context.Context, key string, burstSize int, rate time.Duration) (time.Duration, int, error)
}
