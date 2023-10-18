package azureclient

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis/v2"
	"github.com/go-redis/redis/v8"
)

func BuildRedisClientFromAzure(ctx context.Context, subscriptionID string, resourceGroupName, redisName string) (*redis.Client, error) {
	azRedisClient, err := buildAzRedisClient(subscriptionID, resourceGroupName)
	if err != nil {
		return nil, err
	}

	keys, err := getRedisCred(ctx, azRedisClient, resourceGroupName, redisName)
	if err != nil {
		return nil, err
	}
	redisPassword := *keys.PrimaryKey
	redisHost, err := getRedisHost(ctx, azRedisClient, resourceGroupName, redisName)
	if err != nil {
		return nil, err
	}

	op := &redis.Options{Addr: redisHost, Password: redisPassword, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12}, WriteTimeout: 5 * time.Second}
	client := redis.NewClient(op)
	err = client.Ping(ctx).Err()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to connect with redis instance at %s - %v", redisHost, err))
	}
	return client, nil
}

func buildAzRedisClient(subscriptionID, resourceGroupName string) (*armredis.Client, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	redisClientFactory, err := armredis.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	azRedisClient := redisClientFactory.NewClient()
	return azRedisClient, nil
}

func getRedisCred(ctx context.Context, azRedisClient *armredis.Client, resourceGroupName, redisName string) (*armredis.AccessKeys, error) {
	resp, err := azRedisClient.ListKeys(ctx, resourceGroupName, redisName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.AccessKeys, nil
}

func getRedisHost(ctx context.Context, azRedisClient *armredis.Client, resourceGroupName, redisName string) (string, error) {
	resp, err := azRedisClient.Get(ctx, resourceGroupName, redisName, nil)
	if err != nil {
		return "", err
	}
	hostName := *resp.ResourceInfo.Properties.HostName
	port := *resp.ResourceInfo.Properties.SSLPort
	return fmt.Sprintf("%s:%d", hostName, port), nil
}
