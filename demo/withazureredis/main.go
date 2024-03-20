// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Azure/rate-limiter/demo/handlers"
	"github.com/Azure/rate-limiter/pkg/cache"
	"github.com/Azure/rate-limiter/ratelimiter"
	"github.com/gorilla/mux"
)

const (
	key                          = "billingAccount"
	bucketMaxTokenNumber         = 10
	tokenDropRatePerMin          = 1
	redisDefaultWriteTimeout     = 60 * time.Second
	memoryCacheDefaultExpireTime = 600 * time.Second
	memoryCacheDefaultPurgeTime  = 1200 * time.Second
)

func main() {
	ctx := context.Background()
	redisName := os.Getenv("AZURE_REDIS_NAME")
	if len(redisName) == 0 {
		log.Fatal("AZURE_REDIS_NAME is not set.")
	}

	// example: "c1acf319-6d96-4dfe-b194-c27640869947"
	msiObjectID := os.Getenv("MSI_OBJECT_ID")
	if len(msiObjectID) == 0 {
		log.Fatal("MSI_OBJECT_ID is not set.")
	}

	var uh handlers.ClusterCreateRequestHandlers

	memClient := cache.NewMemCacheClient(memoryCacheDefaultExpireTime, memoryCacheDefaultPurgeTime)

	log.Println("Start to build redis client from azure")
	redisCacheClient, err := cache.NewAzureRedisClient(ctx, fmt.Sprintf("%s.redis.cache.windows.net", redisName), 6380, msiObjectID)
	if err != nil {
		log.Printf("Fail to build redis client from azure, will use in memory cache instead error: %s", err.Error())
		uh = handlers.NewClusterCreateRequestHandlers(ctx, *ratelimiter.NewTokenBucketRateLimiter(memClient, nil), key)
	} else {
		log.Println("Finish build redis client from azure")
		uh = handlers.NewClusterCreateRequestHandlers(ctx, *ratelimiter.NewTokenBucketRateLimiter(memClient, redisCacheClient), key)
	}

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
