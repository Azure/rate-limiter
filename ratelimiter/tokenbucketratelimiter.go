package ratelimiter

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/Azure/rate-limiter/pkg/algorithm"
	"github.com/Azure/rate-limiter/pkg/cache"
)

const (
	tokenNumberKey           = "tokens"
	tokenLastIncreaseTimeKey = "tokenLastIncreaseTime"
)

type TokenBucketRateLimiter struct {
	memCacheClient    cache.CacheClient
	remoteCacheClient cache.CacheClient
}

func NewTokenBucketRateLimiter(memCacheClient, remoteCacheClient cache.CacheClient) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		memCacheClient:    memCacheClient,
		remoteCacheClient: remoteCacheClient,
	}
}

type RateLimiterDecision struct {
	Allowed    bool
	RetryAfter time.Duration
}

// return allow decision and error
func (r *TokenBucketRateLimiter) GetDecision(ctx context.Context, key string, burstSize int, rate time.Duration) (RateLimiterDecision, error) {
	bucket, err := algorithm.NewBucket(rate, burstSize)
	if err != nil {
		// wrong config, fail open
		return RateLimiterDecision{Allowed: true}, err
	}
	// take token from both memcache and remote cache
	allow1, err1 := takeTokenFromCache(ctx, r.remoteCacheClient, bucket, key)
	// memcache won't return any error
	allow2, _ := takeTokenFromCache(ctx, r.memCacheClient, bucket, key)
	if err1 != nil {
		return allow2, err1
	}
	return allow1, nil
}

// return retry after time
func takeTokenFromCache(ctx context.Context, client cache.CacheClient, bucket *algorithm.Bucket, key string) (RateLimiterDecision, error) {
	if client == nil {
		return RateLimiterDecision{Allowed: true}, errors.New("cache client is nil")
	}
	currentCache, err := client.GetCache(ctx, key)
	if err != nil {
		return RateLimiterDecision{Allowed: true}, err
	}
	tokenNumbers, lastIncreaseTime, expireTime, err := bucket.TakeToken(currentCache)
	if err != nil {
		// wrong data
		return RateLimiterDecision{Allowed: true}, err
	}
	if tokenNumbers < 0 {
		// when tokenNumber < 0 means too many requests, return retry after time, 429 and not update cache
		retryAt := lastIncreaseTime.Add(bucket.TokenDropRate)
		return RateLimiterDecision{
			Allowed:    false,
			RetryAfter: time.Until(retryAt),
		}, nil
	}
	err = client.UpdateCache(ctx, key, map[string]string{
		tokenNumberKey:           strconv.Itoa(tokenNumbers),
		tokenLastIncreaseTimeKey: lastIncreaseTime.Format(time.RFC3339),
	}, expireTime)
	if err != nil {
		return RateLimiterDecision{Allowed: true}, err
	}
	return RateLimiterDecision{Allowed: true}, nil
}

func (r *TokenBucketRateLimiter) GetStats(ctx context.Context, key string, burstSize int, rate time.Duration) (int, error) {
	bucket, err := algorithm.NewBucket(rate, burstSize)
	if err != nil {
		// wrong config
		return 0, err
	}
	var currentCache map[string]string
	var err2 error
	if r.remoteCacheClient != nil {
		currentCache, err2 = r.remoteCacheClient.GetCache(ctx, key)
	}
	if r.remoteCacheClient == nil || err2 != nil {
		// use memcache
		currentCache, _ = r.memCacheClient.GetCache(ctx, key)
	}
	if currentCache == nil {
		return burstSize, nil
	}
	return bucket.GetTokenNumber(currentCache)
}
