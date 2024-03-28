package ratelimiter

import "context"

type RateLimiter interface {
	GetDecision(ctx context.Context, key string, burstSize, rate int) (int, error)
}
