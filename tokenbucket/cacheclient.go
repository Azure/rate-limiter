package tokenbucket

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type CacheClient struct {
	ctx    context.Context
	client *redis.Client
}

func NewCacheClient(ctx context.Context, client *redis.Client) *CacheClient {
	return &CacheClient{
		ctx:    ctx,
		client: client,
	}
}

func (c *CacheClient) UpdateCache(key string, cacheData map[string]string, expireTime time.Duration) error {
	_, err := c.client.HSet(c.ctx, key, cacheData).Result()
	if err != nil {
		return err
	}
	c.client.Expire(c.ctx, key, expireTime)
	return nil
}

func (c *CacheClient) GetCache(key string) (map[string]string, error) {
	return c.client.HGetAll(c.ctx, key).Result()
}
