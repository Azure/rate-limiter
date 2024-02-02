package cache

import "time"

type CacheClient interface {
	UpdateCache(key string, cacheData map[string]string, expireTime time.Duration) error
	GetCache(key string) (map[string]string, error)
}
