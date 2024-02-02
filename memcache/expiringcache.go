package memcache

import (
	"container/heap"
	"time"
)

// CacheItem represents an item in the cache.
type CacheItem struct {
	key            int
	value          int
	expirationTime time.Time
}

// ExpiryQueue is a min-heap to keep track of key expirations.
type ExpiryQueue []*CacheItem

func (eq ExpiryQueue) Len() int           { return len(eq) }
func (eq ExpiryQueue) Less(i, j int) bool { return eq[i].expirationTime.Before(eq[j].expirationTime) }
func (eq ExpiryQueue) Swap(i, j int)      { eq[i], eq[j] = eq[j], eq[i] }

// only push the item to the end of the slice, heap.Push will reorder the heap
func (eq *ExpiryQueue) Push(x interface{}) {
	item := x.(*CacheItem)
	*eq = append(*eq, item)
}

// remove the last one, heap.Pop have reordered the heap
func (eq *ExpiryQueue) Pop() interface{} {
	old := *eq
	n := len(old)
	item := old[n-1]
	*eq = old[0 : n-1]
	return item
}

// TimeLimitedCache represents the Time Limited Cache.
type TimeLimitedCache struct {
	cache       map[int]*CacheItem
	expiryQueue ExpiryQueue
}

// NewTimeLimitedCache initializes a new TimeLimitedCache.
func NewTimeLimitedCache() *TimeLimitedCache {
	return &TimeLimitedCache{
		cache:       make(map[int]*CacheItem),
		expiryQueue: make(ExpiryQueue, 0),
	}
}

// set inserts or updates a key-value pair in the cache with a duration.
func (c *TimeLimitedCache) Set(key, value, duration int) bool {
	// Remove expired keys
	c.removeExpiredKeys()

	if item, ok := c.cache[key]; ok {
		// Key already exists, update value and duration
		item.value = value
		item.expirationTime = time.Now().Add(time.Millisecond * time.Duration(duration))
		// 0 is because don't know the index
		heap.Fix(&c.expiryQueue, 0)
		return true
	}

	// Add new key-value pair
	item := &CacheItem{
		key:            key,
		value:          value,
		expirationTime: time.Now().Add(time.Millisecond * time.Duration(duration)),
	}
	heap.Push(&c.expiryQueue, item)
	c.cache[key] = item
	return false
}

// get retrieves the value associated with a key from the cache.
func (c *TimeLimitedCache) Get(key int) int {
	// Remove expired keys
	c.removeExpiredKeys()

	if item, ok := c.cache[key]; ok {
		return item.value
	}
	return -1
}

// removeExpiredKeys removes expired keys from both the cache and the expiryQueue.
func (c *TimeLimitedCache) removeExpiredKeys() {
	currentTime := time.Now()

	// Remove expired keys from both the cache and the expiryQueue
	for len(c.expiryQueue) > 0 && c.expiryQueue[0].expirationTime.Before(currentTime) {
		item := heap.Pop(&c.expiryQueue).(*CacheItem)
		delete(c.cache, item.key)
	}
}
