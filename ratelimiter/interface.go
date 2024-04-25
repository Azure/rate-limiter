package ratelimiter

import (
	"context"
	"time"

	_ "github.com/golang/mock/mockgen/model"
)

//go:generate sh -c "mockgen github.com/Azure/rate-limiter/ratelimiter RateLimiter >./mock_$GOPACKAGE/interface.go"

type RateLimiter interface {
	GetDecision(ctx context.Context, key string, burstSize int, rate time.Duration) (RateLimiterDecision, error)
}
