package ratelimiter

import (
	"context"
	"fmt"
	"net/http"
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

func (r *TokenBucketRateLimiter) GetDecision(ctx context.Context, key string, burstSize, rate int) (int, error) {
	bucket, err := algorithm.NewBucket(rate, burstSize)
	if err != nil {
		// wrong config
		return http.StatusInternalServerError, err
	}
	// take token from both memcache and remote cache
	httpStatusCode1, err1 := takeTokenFromCache(ctx, r.remoteCacheClient, bucket, key)
	// memcache won't return any error
	httpStatusCode2, _ := takeTokenFromCache(ctx, r.memCacheClient, bucket, key)
	if httpStatusCode1 != http.StatusInternalServerError {
		return httpStatusCode1, err1
	}
	return httpStatusCode2, nil
}

func takeTokenFromCache(ctx context.Context, client cache.CacheClient, bucket *algorithm.Bucket, key string) (int, error) {
	if client == nil {
		return http.StatusInternalServerError, fmt.Errorf("cache client is nil")
	}
	currentCache, err := client.GetCache(ctx, key)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	tokenNumbers, lastIncreaseTime, expireData, err := bucket.TakeToken(currentCache)
	if err != nil {
		// wrong data
		return http.StatusInternalServerError, err
	}
	if tokenNumbers < 0 {
		// when tokenNumber < 0 means too many requests, return 429 and not update cache
		return http.StatusTooManyRequests, nil
	}
	_ = client.UpdateCache(ctx, key, map[string]string{
		tokenNumberKey:           strconv.Itoa(tokenNumbers),
		tokenLastIncreaseTimeKey: lastIncreaseTime.Format(time.RFC3339),
	}, expireData)
	return http.StatusOK, nil
}

func (r *TokenBucketRateLimiter) GetStats(ctx context.Context, key string, burstSize, rate int) (int, error) {
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
