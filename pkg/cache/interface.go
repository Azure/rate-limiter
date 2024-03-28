package cache

import (
	"context"
	"time"
)

type CacheClient interface {
	UpdateCache(ctx context.Context, key string, cacheData map[string]string, expireTime time.Duration) error
	GetCache(ctx context.Context, key string) (map[string]string, error)
}
