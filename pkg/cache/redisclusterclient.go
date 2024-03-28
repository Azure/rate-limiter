package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClusterCacheClient struct {
	client *redis.ClusterClient
}

func NewClusterClient(client *redis.ClusterClient) *RedisClusterCacheClient {
	return &RedisClusterCacheClient{
		client: client,
	}
}

func (c *RedisClusterCacheClient) UpdateCache(ctx context.Context, key string, cacheData map[string]string, expireTime time.Duration) error {
	_, err := c.client.HSet(ctx, key, cacheData).Result()
	if err != nil {
		return err
	}
	c.client.Expire(ctx, key, expireTime)
	return nil
}

func (c *RedisClusterCacheClient) GetCache(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

func (c *RedisClusterCacheClient) GetMemoryUsage(ctx context.Context, key string) (int64, error) {
	return c.client.MemoryUsage(ctx, key).Result()
}
