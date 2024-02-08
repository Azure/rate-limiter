package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	ctx    context.Context
	client *redis.Client
}

func NewRedisClient(ctx context.Context, client *redis.Client) *RedisClient {
	return &RedisClient{
		ctx:    ctx,
		client: client,
	}
}

func (c *RedisClient) UpdateCache(key string, cacheData map[string]string, expireTime time.Duration) error {
	_, err := c.client.HSet(c.ctx, key, cacheData).Result()
	if err != nil {
		return err
	}
	c.client.Expire(c.ctx, key, expireTime)
	return nil
}

func (c *RedisClient) GetCache(key string) (map[string]string, error) {
	return c.client.HGetAll(c.ctx, key).Result()
}

func (c *RedisClient) GetMemoryUsage(key string) (int64, error) {
	return c.client.MemoryUsage(c.ctx, key).Result()
}
