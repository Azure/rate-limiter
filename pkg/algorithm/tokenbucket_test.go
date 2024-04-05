package algorithm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReconstructTokenStateFromCache(t *testing.T) {
	bucket, err := NewBucket(30*time.Second, 10)
	assert.Nil(t, err)
	// no need to change token last increase time
	lastIncreaseTime := time.Now().Add(-time.Second * 10).Format(time.RFC3339)
	currentCache := map[string]string{
		tokenNumberKey:           "5",
		tokenLastIncreaseTimeKey: lastIncreaseTime,
	}
	bucketStats, err := bucket.reconstructTokenStateFromCache(currentCache)
	assert.Nil(t, err)
	assert.Equal(t, 5, bucketStats.tokenNumbers)
	assert.Equal(t, lastIncreaseTime, bucketStats.lastIncreaseTime.Format(time.RFC3339))

	currentCache = map[string]string{
		tokenNumberKey:           "5",
		tokenLastIncreaseTimeKey: time.Now().Add(-time.Minute).Format(time.RFC3339),
	}

	bucketStats, err = bucket.reconstructTokenStateFromCache(currentCache)
	assert.Nil(t, err)
	assert.Equal(t, 7, bucketStats.tokenNumbers)
	// last increase time should be around now, roughly check if it's within 1 seconds
	assert.True(t, bucketStats.lastIncreaseTime.After(time.Now().Add(-time.Second*1)))
}

func TestReconstructTokenStateFromCacheOverBurstSize(t *testing.T) {
	bucket, err := NewBucket(30*time.Second, 10)
	assert.Nil(t, err)
	currentCache := map[string]string{
		tokenNumberKey:           "5",
		tokenLastIncreaseTimeKey: time.Now().Add(-3 * time.Minute).Format(time.RFC3339),
	}

	bucketStats, err := bucket.reconstructTokenStateFromCache(currentCache)
	assert.Nil(t, err)
	assert.Equal(t, 10, bucketStats.tokenNumbers)
	// last increase time should be around now, roughly check if it's within 1 seconds
	assert.True(t, bucketStats.lastIncreaseTime.After(time.Now().Add(-time.Second*1)))
}

func TestReconstructTokenStateFromCacheWithWrongData(t *testing.T) {
	bucket, err := NewBucket(30*time.Second, 10)
	assert.Nil(t, err)
	currentCache := map[string]string{
		tokenNumberKey:           "5",
		tokenLastIncreaseTimeKey: "wrong time format",
	}

	_, err = bucket.reconstructTokenStateFromCache(currentCache)
	assert.NotNil(t, err)

	currentCache = map[string]string{
		tokenNumberKey:           "-3",
		tokenLastIncreaseTimeKey: time.Now().Add(-time.Minute).Format(time.RFC3339),
	}
	_, err = bucket.reconstructTokenStateFromCache(currentCache)
	assert.NotNil(t, err)
}

func TestTakeToken(t *testing.T) {
	bucket, err := NewBucket(30*time.Second, 10)
	assert.Nil(t, err)
	currentCache := map[string]string{
		tokenNumberKey:           "5",
		tokenLastIncreaseTimeKey: time.Now().Add(-time.Minute).Format(time.RFC3339),
	}

	tokenNumbers, lastIncreaseTime, expireTime, err := bucket.TakeToken(currentCache)
	assert.Nil(t, err)
	assert.Equal(t, 6, tokenNumbers)
	// about now
	assert.True(t, lastIncreaseTime.After(time.Now().Add(-time.Second*1)))
	// about 120s before expire
	assert.Equal(t, 1, int(120*time.Second/expireTime))
}

func TestTakeTokenWrongData(t *testing.T) {
	bucket, err := NewBucket(30*time.Second, 10)
	assert.Nil(t, err)
	currentCache := map[string]string{
		tokenNumberKey:           "wrong-data",
		tokenLastIncreaseTimeKey: time.Now().Add(-time.Minute).Format(time.RFC3339),
	}

	tokenNumbers, lastIncreaseTime, expireTime, err := bucket.TakeToken(currentCache)
	assert.NotNil(t, err)
	assert.Equal(t, 9, tokenNumbers)
	// about now
	assert.True(t, lastIncreaseTime.After(time.Now().Add(-time.Second*1)))
	// about 30s before expire
	assert.Equal(t, 1, int(30*time.Second/expireTime))
}
