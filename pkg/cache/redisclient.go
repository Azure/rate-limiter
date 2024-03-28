package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(ctx context.Context, client *redis.Client) *RedisClient {
	return &RedisClient{
		client: client,
	}
}

func (c *RedisClient) UpdateCache(ctx context.Context, key string, cacheData map[string]string, expireTime time.Duration) error {
	err := c.client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to connect with redis instance: %s", err.Error())
	}
	_, err = c.client.HSet(ctx, key, cacheData).Result()
	if err != nil {
		return err
	}
	c.client.Expire(ctx, key, expireTime)
	return nil
}

func (c *RedisClient) GetCache(ctx context.Context, key string) (map[string]string, error) {
	err := c.client.Ping(ctx).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to connect with redis instance: %s", err.Error())
	}
	return c.client.HGetAll(ctx, key).Result()
}

func (c *RedisClient) GetMemoryUsage(ctx context.Context, key string) (int64, error) {
	err := c.client.Ping(ctx).Err()
	if err != nil {
		return 0, fmt.Errorf("failed to connect with redis instance: %s", err.Error())
	}
	return c.client.MemoryUsage(ctx, key).Result()
}
