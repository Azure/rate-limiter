package tokenbucket

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type tokenBucket struct {
	tokenNumbers          int
	tokenLastIncreaseTime time.Time
}

const (
	bucketMaxTokenNumber = 10
	tokenDropRatePerMin  = 1
)

// Self maintained token bucket
func reconstructTokenBucketFromMap(currentCache map[string]string) tokenBucket {
	if len(currentCache) == 0 {
		return tokenBucket{tokenNumbers: bucketMaxTokenNumber, tokenLastIncreaseTime: time.Now()}
	}
	lastSavedTokens, _ := strconv.Atoi(currentCache["tokens"])
	tokenLastIncreaseTime, _ := time.Parse(time.RFC3339, currentCache["tokenLastIncreaseTime"])
	currentTime := time.Now()
	elapsedTime := currentTime.Sub(tokenLastIncreaseTime)
	fmt.Printf("before calculation: lastSavedTokens: %d, savedLastUpdatedTime: %s, timeNow: %s, elapsedTime: %s\n", lastSavedTokens, tokenLastIncreaseTime, currentTime, elapsedTime)
	// calculate tokens
	shouldIncreaseTokens := int(math.Floor(elapsedTime.Seconds()/60)) * tokenDropRatePerMin
	tokensNow := shouldIncreaseTokens + lastSavedTokens
	if tokensNow > bucketMaxTokenNumber {
		tokensNow = bucketMaxTokenNumber
	}
	if shouldIncreaseTokens > 0 {
		// update tokenLastIncreaseTime if tokens are increased
		tokenLastIncreaseTime = tokenLastIncreaseTime.Add(time.Duration(shouldIncreaseTokens) * time.Minute)
	}
	fmt.Printf("after calculation: tokensNow: %d, tokenLastIncreaseTime: %s\n", tokensNow, tokenLastIncreaseTime)
	return tokenBucket{tokenNumbers: tokensNow, tokenLastIncreaseTime: tokenLastIncreaseTime}
}

func UpdateBucketInCache(ctx context.Context, client *redis.Client, key string, currentCache map[string]string) (int, error) {
	tokenBucket := reconstructTokenBucketFromMap(currentCache)

	if tokenBucket.tokenNumbers <= 0 {
		return http.StatusTooManyRequests, errors.New(fmt.Sprintf("too many requests from billing account: %s", key))
	}
	tokenBucket.tokenNumbers--

	_, err := client.HSet(ctx, key, map[string]string{
		"tokens":                strconv.Itoa(tokenBucket.tokenNumbers),
		"tokenLastIncreaseTime": tokenBucket.tokenLastIncreaseTime.Format(time.RFC3339),
	}).Result()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	fmt.Printf("set data for: %s\n", key)

	// let's say there is currently 6 tokens, max token number is 10, token drop rate is 1 per minute
	// 4 minutes later, the bucket will reach max token number
	// then we don't need to keep this bucket in the cache
	// when user request with this billing account id come again after expiration, we will just start a new bucket with 10 tokens
	timeForCurrentbucketToReachMaxTokenNumber := tokenBucket.tokenLastIncreaseTime.Add(time.Duration(math.Ceil(float64(bucketMaxTokenNumber-tokenBucket.tokenNumbers)/tokenDropRatePerMin)) * time.Minute)
	client.Expire(ctx, key, time.Until(timeForCurrentbucketToReachMaxTokenNumber))
	fmt.Printf("set data expire time to %s for: %s, expire duration %f minutes\n", timeForCurrentbucketToReachMaxTokenNumber, key, time.Until(timeForCurrentbucketToReachMaxTokenNumber).Minutes())
	return http.StatusOK, nil
}
