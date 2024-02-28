// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pkg/cache"
	"ratelimiter"
	"test/handlers"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

const (
	key                  = "billingAccount"
	bucketMaxTokenNumber = 10
	tokenDropRatePerMin  = 1
)

func BuildRedisClientFromAzure(ctx context.Context, subscriptionID string, resourceGroupName, redisName string) (*redis.Client, error) {
	azRedisClient, err := buildAzRedisClientWithLocalAuth(subscriptionID, resourceGroupName)
	if err != nil {
		return nil, err
	}

	resp, err := azRedisClient.ListKeys(ctx, resourceGroupName, redisName, nil)
	if err != nil {
		return nil, err
	}
	redisPassword := *resp.AccessKeys.PrimaryKey

	keyResp, err := azRedisClient.Get(ctx, resourceGroupName, redisName, nil)
	if err != nil {
		return nil, err
	}
	hostName := *keyResp.ResourceInfo.Properties.HostName
	port := *keyResp.ResourceInfo.Properties.SSLPort
	redisHost := fmt.Sprintf("%s:%d", hostName, port)

	return buildRedisClient(ctx, redisHost, redisPassword)
}

func buildRedisClient(ctx context.Context, redisHost, redisPassword string) (*redis.Client, error) {
	op := &redis.Options{Addr: redisHost, Password: redisPassword, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12}, WriteTimeout: 60 * time.Second}
	client := redis.NewClient(op)
	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to connect with redis instance at %s - %v", redisHost, err))
	}
	return client, nil
}

func buildAzRedisClientWithLocalAuth(subscriptionID, resourceGroupName string) (*armredis.Client, error) {
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

func buildAzRedisClientWithMSI(subscriptionID, resourceGroupName, msiResourceId string) (*armredis.Client, error) {
	id := azidentity.ResourceID(msiResourceId)
	opts := &azidentity.ManagedIdentityCredentialOptions{ID: id}
	cred, err := azidentity.NewManagedIdentityCredential(opts)
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

func main() {
	ctx := context.Background()

	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}
	resourceGroupName := os.Getenv("AZURE_RESOURCE_GROUP")
	if len(resourceGroupName) == 0 {
		log.Fatal("AZURE_RESOURCE_GROUP is not set.")
	}
	redisName := os.Getenv("AZURE_REDIS_NAME")
	if len(redisName) == 0 {
		log.Fatal("AZURE_REDIS_NAME is not set.")
	}

	redisClient, err := BuildRedisClientFromAzure(ctx, subscriptionID, resourceGroupName, redisName)
	if err != nil {
		log.Fatal(err)
	}

	cacheClient := cache.NewRedisClient(ctx, redisClient)

	memClient := cache.NewMemCacheClient(10*time.Minute, 20*time.Minute)

	uh := handlers.NewClusterCreateRequestHandlers(ctx, *ratelimiter.NewTokenBucketRateLimiter(memClient, cacheClient), key)

	router := mux.NewRouter()
	router.HandleFunc(fmt.Sprintf("/%s/", key), uh.HandleRequest).Methods(http.MethodPost)
	router.HandleFunc(fmt.Sprintf("/%s/{%s}", key, key), uh.GetBucketStats).Methods(http.MethodGet)

	log.Println("Start server on port 8080")

	server := http.Server{Addr: ":8080", Handler: router}
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Server start to listen on port 8080")
	}()
	// listening to OS shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Println("Got shutdown signal, shutting down server gracefully...")
	server.Shutdown(ctx)
}
