package ratelimiter

type RateLimiter interface {
	GetDecision(key string, burstSize, rate int) (int, error)
}
