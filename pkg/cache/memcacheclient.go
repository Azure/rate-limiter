package cache

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
)

type MemCacheClient struct {
	memCache *cache.Cache
}

func NewMemCacheClient(defaultExpireTime, defaultPurgeTime time.Duration) *MemCacheClient {
	return &MemCacheClient{
		memCache: cache.New(defaultExpireTime, defaultPurgeTime),
	}
}

func (c MemCacheClient) UpdateCache(ctx context.Context, key string, cacheData map[string]string, expireTime time.Duration) error {
	c.memCache.Set(key, cacheData, expireTime)
	return nil
}

func (c MemCacheClient) GetCache(ctx context.Context, key string) (map[string]string, error) {
	cacheData, found := c.memCache.Get(key)
	if !found {
		return nil, nil
	}
	return cacheData.(map[string]string), nil

}
