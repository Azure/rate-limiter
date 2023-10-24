package tokenbucket

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClusterCacheClient struct {
	ctx    context.Context
	client *redis.ClusterClient
}

func NewClusterClient(ctx context.Context, client *redis.ClusterClient) *RedisClusterCacheClient {
	return &RedisClusterCacheClient{
		ctx:    ctx,
		client: client,
	}
}

func (c *RedisClusterCacheClient) UpdateCache(key string, cacheData map[string]string, expireTime time.Duration) error {
	_, err := c.client.HSet(c.ctx, key, cacheData).Result()
	if err != nil {
		return err
	}
	c.client.Expire(c.ctx, key, expireTime)
	return nil
}

func (c *RedisClusterCacheClient) GetCache(key string) (map[string]string, error) {
	return c.client.HGetAll(c.ctx, key).Result()
}

func (c *RedisClusterCacheClient) GetMemoryUsage(key string)(int64, error) {
	return c.client.MemoryUsage(c.ctx, key).Result()
}