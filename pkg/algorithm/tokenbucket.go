package algorithm

import (
	"errors"
	"strconv"
	"time"
)

type Bucket struct {
	TokenDropRate time.Duration // add a token to bucket every tokenDropRate
	BurstSize     int
}

type tokenState struct {
	tokenNumbers int
	// LastIncreaseTime is the last time the bucket's tokens number increase
	lastIncreaseTime time.Time
}

const (
	tokenNumberKey           = "tokens"
	tokenLastIncreaseTimeKey = "tokenLastIncreaseTime"
	DefaultTokenDropRate     = time.Minute
	DefaultBurstSize         = 10
)

func NewBucket(tokenDropRate time.Duration, burstSize int) (*Bucket, error) {
	if burstSize <= 0 {
		return nil, errors.New("burst size must be greater than 0")
	}
	bucket := &Bucket{
		TokenDropRate: tokenDropRate,
		BurstSize:     burstSize,
	}
	return bucket, nil
}

// return bucket current token number, last time token increase, bucket expire time, error
// bucket expire time = time when the bucket is full(reach burst size)
// let's say there is currently 6 tokens, max token number is 10, token drop every minute
// 4 minutes later, the bucket will reach max token number
// then we don't need to keep this bucket in the cache
// when user request with this id come again after expiration, we will just start a new bucket with 10 tokens
func (b *Bucket) TakeToken(currentCache map[string]string) (int, time.Time, time.Duration, error) {
	var tokenState *tokenState
	var err error
	if tokenState, err = b.reconstructTokenStateFromCache(currentCache); err != nil {
		return 0, time.Now(), time.Millisecond, err
	}

	tokenState.tokenNumbers--

	var tokesLeftForBucketToFull int
	if tokenState.tokenNumbers < 0 {
		tokesLeftForBucketToFull = b.BurstSize
	} else {
		tokesLeftForBucketToFull = b.BurstSize - tokenState.tokenNumbers
	}

	timeForCurrentbucketToFull := tokenState.lastIncreaseTime.Add(time.Duration(tokesLeftForBucketToFull) * b.TokenDropRate)
	now := time.Now()
	expireTime := timeForCurrentbucketToFull.Sub(now)
	return tokenState.tokenNumbers, tokenState.lastIncreaseTime, expireTime, nil
}

func (b *Bucket) GetTokenNumber(currentCache map[string]string) (int, error) {
	var tokenState *tokenState
	var err error
	if tokenState, err = b.reconstructTokenStateFromCache(currentCache); err != nil {
		return 0, err
	}
	return tokenState.tokenNumbers, nil
}

// cache record: current token number, last time token increase
// why record a time stamp for last time token increase? - to reconstruct token numbers
// let's say burst size is 10, token drop every minute
// record for this bucket in cache: token number: 0, last increase time: 10:00:00
// at 10:00:30, this bucket is: token number 0, last increase time: 10:00:00
// at 10:01:30, this bucket is: token number 1, last increase time: 10:01:00
// at 10:06:20, this bucket is: token number 6, last increase time: 10:06:00
// at 10:30:30, this bucket is: token number 10(burst size), last increase time: 10:30:00
func (b *Bucket) reconstructTokenStateFromCache(currentCache map[string]string) (*tokenState, error) {
	if len(currentCache) == 0 {
		tokenState := tokenState{tokenNumbers: b.BurstSize, lastIncreaseTime: time.Now()}
		return &tokenState, nil
	}
	lastSavedTokens, err := strconv.Atoi(currentCache[tokenNumberKey])
	if err != nil {
		return nil, err
	}
	if lastSavedTokens < 0 {
		return nil, errors.New("wrong token number")
	}
	tokenLastIncreaseTime, err := time.Parse(time.RFC3339, currentCache[tokenLastIncreaseTimeKey])
	if err != nil {
		return nil, err
	}
	currentTime := time.Now()
	elapsedTime := currentTime.Sub(tokenLastIncreaseTime)
	// calculate tokens
	shouldIncreaseTokens := int(elapsedTime / b.TokenDropRate)
	tokensNow := shouldIncreaseTokens + lastSavedTokens
	if tokensNow > b.BurstSize {
		tokensNow = b.BurstSize
	}
	if shouldIncreaseTokens > 0 {
		// update tokenLastIncreaseTime if tokens are increased
		tokenLastIncreaseTime = tokenLastIncreaseTime.Add(time.Duration(shouldIncreaseTokens) * b.TokenDropRate)
	}
	tokenState := tokenState{tokenNumbers: tokensNow, lastIncreaseTime: tokenLastIncreaseTime}
	return &tokenState, nil
}
