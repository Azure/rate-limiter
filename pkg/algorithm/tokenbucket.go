package algorithm

import (
	"errors"
	"math"
	"strconv"
	"time"
)

type Bucket struct {
	tokenDropRatePerMin int
	burstSize           int
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

func NewBucket(tokenDropRatePerMin, burstSize int) (*Bucket, error) {
	if tokenDropRatePerMin <= 0 {
		return nil, errors.New("token drop rate per minute must be greater than 0")
	}
	if burstSize <= 0 {
		return nil, errors.New("burst size must be greater than 0")
	}
	bucket := &Bucket{
		tokenDropRatePerMin: tokenDropRatePerMin,
		burstSize:           burstSize,
	}
	return bucket, nil
}

func (b *Bucket) TakeToken(currentCache map[string]string) (int, time.Time, time.Duration, error) {
	var tokenState *tokenState
	var err error
	if tokenState, err = b.reconstructTokenStateFromCache(currentCache); err != nil {
		return 0, time.Now(), time.Millisecond, err
	}

	tokenState.tokenNumbers--

	// let's say there is currently 6 tokens, max token number is 10, token drop rate is 1 per minute
	// 4 minutes later, the bucket will reach max token number
	// then we don't need to keep this bucket in the cache
	// when user request with this billing account id come again after expiration, we will just start a new bucket with 10 tokens
	tokesLeftForBucketToFull := b.burstSize - tokenState.tokenNumbers
	timeForCurrentbucketToFull := tokenState.lastIncreaseTime.Add(time.Duration(math.Ceil(float64(tokesLeftForBucketToFull)/float64(b.tokenDropRatePerMin))) * time.Minute)

	return tokenState.tokenNumbers, tokenState.lastIncreaseTime, time.Until(timeForCurrentbucketToFull), nil
}

func (b *Bucket) GetTokenNumber(currentCache map[string]string) (int, error) {
	var tokenState *tokenState
	var err error
	if tokenState, err = b.reconstructTokenStateFromCache(currentCache); err != nil {
		return 0, err
	}
	return tokenState.tokenNumbers, nil
}

func (b *Bucket) reconstructTokenStateFromCache(currentCache map[string]string) (*tokenState, error) {
	if len(currentCache) == 0 {
		tokenState := tokenState{tokenNumbers: b.burstSize, lastIncreaseTime: time.Now()}
		return &tokenState, nil
	}
	lastSavedTokens, err := strconv.Atoi(currentCache[tokenNumberKey])
	if err != nil {
		return nil, err
	}
	tokenLastIncreaseTime, err := time.Parse(time.RFC3339, currentCache[tokenLastIncreaseTimeKey])
	if err != nil {
		return nil, err
	}
	currentTime := time.Now()
	elapsedTime := currentTime.Sub(tokenLastIncreaseTime)
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
	tokenState := tokenState{tokenNumbers: tokensNow, lastIncreaseTime: tokenLastIncreaseTime}
	return &tokenState, nil
}
