package tokenbucket

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"go.goms.io/rate-limiter-backed-by-redis-cache/cache"
)

type Bucket struct {
	key                 string
	client              cache.CacheClient
	tokenDropRatePerMin int
	burstSize           int
	tokenState          tokenState
}

type tokenState struct {
	tokenNumbers int
	// LastIncreaseTime is the last time the bucket's tokens number increase
	lastIncreaseTime time.Time
}

const (
	tokenNumberKey             = "tokens"
	tokenLastIncreaseTimeKey   = "tokenLastIncreaseTime"
	DefaultTokenDropRatePerMin = 1
	DefaultBurstSize           = 10
)

func NewBucket(ctx context.Context, cacheClient cache.CacheClient, key string, tokenDropRatePerMin, burstSize int) (*Bucket, error) {
	if tokenDropRatePerMin <= 0 {
		return nil, errors.New("token drop rate per minute must be greater than 0")
	}
	if burstSize <= 0 {
		return nil, errors.New("burst size must be greater than 0")
	}
	bucket := &Bucket{
		key:                 key,
		client:              cacheClient,
		tokenDropRatePerMin: tokenDropRatePerMin,
		burstSize:           burstSize,
	}
	err := bucket.initTokenState(ctx)
	if err != nil {
		return nil, err
	}
	return bucket, nil
}

func (b *Bucket) Allow() bool {
	return b.tokenState.tokenNumbers > 0
}

func (b *Bucket) TakeToken() (int, error) {
	if b.tokenState.tokenNumbers <= 0 {
		return http.StatusTooManyRequests, errors.New(fmt.Sprintf("too many requests from billing account: %s", b.key))
	}
	b.tokenState.tokenNumbers--

	fmt.Printf("set data for: %s\n", b.key)

	// let's say there is currently 6 tokens, max token number is 10, token drop rate is 1 per minute
	// 4 minutes later, the bucket will reach max token number
	// then we don't need to keep this bucket in the cache
	// when user request with this billing account id come again after expiration, we will just start a new bucket with 10 tokens
	tokesLeftForBucketToFull := b.burstSize - b.tokenState.tokenNumbers
	timeForCurrentbucketToFull := b.tokenState.lastIncreaseTime.Add(time.Duration(math.Ceil(float64(tokesLeftForBucketToFull)/float64(b.tokenDropRatePerMin))) * time.Minute)
	err := b.client.UpdateCache(b.key, map[string]string{
		tokenNumberKey:           strconv.Itoa(b.tokenState.tokenNumbers),
		tokenLastIncreaseTimeKey: b.tokenState.lastIncreaseTime.Format(time.RFC3339),
	}, time.Until(timeForCurrentbucketToFull))
	if err != nil {
		return http.StatusInternalServerError, err
	}
	fmt.Printf("set data expire time to %s for: %s, expire duration %f minutes\n", timeForCurrentbucketToFull, b.key, time.Until(timeForCurrentbucketToFull).Minutes())
	memoryUsage, err := b.client.GetMemoryUsage(b.key)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	fmt.Printf("memory usage for %s: %d bytes\n", b.key, memoryUsage)
	return http.StatusOK, nil
}

func (b *Bucket) GetBucketStats() (int, int, int) {
	return b.tokenState.tokenNumbers, b.burstSize, b.tokenDropRatePerMin
}

func (b *Bucket) initTokenState(ctx context.Context) error {
	currentCache, err := b.client.GetCache(b.key)
	if err != nil {
		return err
	}
	return b.reconstructTokenStateFromCache(currentCache)
}

func (b *Bucket) reconstructTokenStateFromCache(currentCache map[string]string) error {
	if len(currentCache) == 0 {
		b.tokenState = tokenState{tokenNumbers: b.burstSize, lastIncreaseTime: time.Now()}
		return nil
	}
	lastSavedTokens, err := strconv.Atoi(currentCache[tokenNumberKey])
	if err != nil {
		return err
	}
	tokenLastIncreaseTime, err := time.Parse(time.RFC3339, currentCache[tokenLastIncreaseTimeKey])
	if err != nil {
		return err
	}
	currentTime := time.Now()
	elapsedTime := currentTime.Sub(tokenLastIncreaseTime)
	fmt.Printf("before calculation: lastSavedTokens: %d, savedLastUpdatedTime: %s, timeNow: %s, elapsedTime: %s\n", lastSavedTokens, tokenLastIncreaseTime, currentTime, elapsedTime)
	// calculate tokens
	shouldIncreaseTokens := int(math.Floor(elapsedTime.Seconds()/60)) * b.tokenDropRatePerMin
	tokensNow := shouldIncreaseTokens + lastSavedTokens
	if tokensNow > b.burstSize {
		tokensNow = b.burstSize
	}
	if shouldIncreaseTokens > 0 {
		// update tokenLastIncreaseTime if tokens are increased
		tokenLastIncreaseTime = tokenLastIncreaseTime.Add(time.Duration(shouldIncreaseTokens) * time.Minute)
	}
	fmt.Printf("after calculation: tokensNow: %d, tokenLastIncreaseTime: %s\n", tokensNow, tokenLastIncreaseTime)
	b.tokenState = tokenState{tokenNumbers: tokensNow, lastIncreaseTime: tokenLastIncreaseTime}
	return nil
}
