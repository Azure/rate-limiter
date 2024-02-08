package ratelimiter

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"pkg/cache"
	"pkg/tokenbucket"
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

func (r *TokenBucketRateLimiter) GetDecision(key string, burstSize, rate int) (int, error) {
	bucket, err := tokenbucket.NewBucket(rate, burstSize)
	if err != nil {
		// wrong config
		return http.StatusInternalServerError, err
	}

	currentCache, err := r.remoteCacheClient.GetCache(key)
	if err != nil {
		// use memcache
		currentCache, _ = r.memCacheClient.GetCache(key)
		tokenNumbers, lastIncreaseTime, expireData, err := bucket.TakeToken(currentCache)
		if err != nil {
			// wrong data
			return http.StatusInternalServerError, err
		}
		if tokenNumbers < 0 {
			return http.StatusTooManyRequests, errors.New(fmt.Sprintf("too many requests from: %s", key))
		}
		_ = r.memCacheClient.UpdateCache(key, map[string]string{
			tokenNumberKey:           strconv.Itoa(tokenNumbers),
			tokenLastIncreaseTimeKey: lastIncreaseTime.Format(time.RFC3339),
		}, expireData)
		return http.StatusOK, nil
	}
	tokenNumbers, lastIncreaseTime, expireData, err := bucket.TakeToken(currentCache)
	if err != nil {
		// wrong data
		return http.StatusInternalServerError, err
	}
	if tokenNumbers < 0 {
		return http.StatusTooManyRequests, errors.New(fmt.Sprintf("too many requests from: %s", key))
	}
	_ = r.remoteCacheClient.UpdateCache(key, map[string]string{
		tokenNumberKey:           strconv.Itoa(tokenNumbers),
		tokenLastIncreaseTimeKey: lastIncreaseTime.Format(time.RFC3339),
	}, expireData)
	return http.StatusOK, nil
}

func (r *TokenBucketRateLimiter) GetStats(key string, burstSize, rate int) (int, error) {
	bucket, err := tokenbucket.NewBucket(rate, burstSize)
	if err != nil {
		// wrong config
		return http.StatusInternalServerError, err
	}
	var currentCache map[string]string
	currentCache, err = r.remoteCacheClient.GetCache(key)
	if err != nil {
		// use memcache
		currentCache, _ = r.memCacheClient.GetCache(key)
	}
	return bucket.GetTokenNumber(currentCache)
}
