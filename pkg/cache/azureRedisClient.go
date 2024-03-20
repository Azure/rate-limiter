package cache

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/go-redis/redis/v8"
)

const (
	redisDefaultWriteTimeout     = 60 * time.Second
	memoryCacheDefaultExpireTime = 600 * time.Second
	memoryCacheDefaultPurgeTime  = 1200 * time.Second
	azureRedisScope              = "https://redis.azure.com/.default"
)

var (
	expiringWindow = time.Second * 30
)

type AzureRedisClient struct {
	ctx          context.Context
	redisClient  *redis.Client
	tokenFetcher *azureCacheTokenFetcher
}

type azureCacheTokenFetcher struct {
	ctx         context.Context
	accessToken azcore.AccessToken
	cred        *azidentity.DefaultAzureCredential
}

func NewAzureCacheTokenFetcherWithMSI(ctx context.Context) (*azureCacheTokenFetcher, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	return &azureCacheTokenFetcher{
		ctx:  ctx,
		cred: cred,
	}, err
}

// getToken gets a new token if token is expired
func (d *azureCacheTokenFetcher) getToken() (string, error) {
	if d.tokenExpired() {
		if err := d.refreshToken(d.ctx); err != nil {
			return "", fmt.Errorf("refress token err: %w", err)
		}
	}
	return d.accessToken.Token, nil
}

func (d *azureCacheTokenFetcher) refreshToken(ctx context.Context) error {
	token, err := d.cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{azureRedisScope}})
	if err != nil {
		return fmt.Errorf("get token err: %w", err)
	}
	d.accessToken = token
	return nil
}

func (d *azureCacheTokenFetcher) tokenExpired() bool {
	if d.accessToken.ExpiresOn.IsZero() || d.accessToken.ExpiresOn.After(time.Now().Add(-expiringWindow)) {
		return true
	}
	return false
}

func NewAzureRedisClient(ctx context.Context, hostName string, port int, identityObjectID string) (*AzureRedisClient, error) {
	tokenFetcher, err := NewAzureCacheTokenFetcherWithMSI(ctx)
	if err != nil {
		return nil, err
	}
	redisPassword, err := tokenFetcher.getToken()
	if err != nil {
		return nil, err
	}
	redisClient, err := buildRedisClient(ctx, fmt.Sprintf("%s:%d", hostName, port), identityObjectID, redisPassword)
	if err != nil {
		return nil, err
	}
	return &AzureRedisClient{
		ctx:          ctx,
		redisClient:  redisClient,
		tokenFetcher: tokenFetcher,
	}, nil
}

func buildRedisClient(ctx context.Context, redisHost, msiObjectID, redisPassword string) (*redis.Client, error) {
	op := &redis.Options{
		Addr:         redisHost,
		Username:     msiObjectID,
		Password:     redisPassword,
		TLSConfig:    &tls.Config{MinVersion: tls.VersionTLS12},
		WriteTimeout: redisDefaultWriteTimeout,
	}
	client := redis.NewClient(op)
	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to connect with redis instance at %s - %v", redisHost, err))
	}
	return client, nil
}

func (c *AzureRedisClient) UpdateCache(key string, cacheData map[string]string, expireTime time.Duration) error {
	if c.tokenFetcher.tokenExpired() {
		if err := c.tokenFetcher.refreshToken(c.ctx); err != nil {
			return err
		}
		// update password when token expired
		newClient, err := buildRedisClient(c.ctx, c.redisClient.Options().Addr, c.redisClient.Options().Username, c.tokenFetcher.accessToken.Token)
		if err != nil {
			return err
		}
		c.redisClient = newClient
	}
	err := c.redisClient.Ping(c.ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to connect with redis instance: %s", err.Error())
	}
	_, err = c.redisClient.HSet(c.ctx, key, cacheData).Result()
	if err != nil {
		return err
	}
	c.redisClient.Expire(c.ctx, key, expireTime)
	return nil
}

func (c *AzureRedisClient) GetCache(key string) (map[string]string, error) {
	err := c.redisClient.Ping(c.ctx).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to connect with redis instance: %s", err.Error())
	}
	return c.redisClient.HGetAll(c.ctx, key).Result()
}
