// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/mux"
	"pkg/cache"
	"test/handlers"
)

const (
	key                  = "billingAccount"
	bucketMaxTokenNumber = 10
	tokenDropRatePerMin  = 1
)

func BuildRedisClusterClient(ctx context.Context, redisHost, redisPassword string) (*redis.ClusterClient, error) {
	var op *redis.ClusterOptions
	if len(redisPassword) == 0 {
		op = &redis.ClusterOptions{Addrs: strings.Split(redisHost, ",")}
	} else {
		op = &redis.ClusterOptions{Addrs: strings.Split(redisHost, ","), Password: redisPassword}
	}
	client := redis.NewClusterClient(op)
	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to connect with redis instance at %s - %v", redisHost, err))
	}
	return client, nil
}

func main() {
	ctx := context.Background()

	redisHost := os.Getenv("REDIS_HOST")
	if len(redisHost) == 0 {
		log.Fatal("REDIS_HOST is not set.")
	}
	// connect to redis cluster
	redisClusterClient, err := cache.BuildRedisClusterClient(ctx, redisHost, os.Getenv("REDIS_PASSWORD"))
	if err != nil {
		log.Fatal(err)
	}

	uh := handlers.NewClusterCreateRequestHandlers(ctx, cache.NewClusterClient(ctx, redisClusterClient), key)

	router := mux.NewRouter()
	router.HandleFunc(fmt.Sprintf("/%s/", key), uh.HandleRequest).Methods(http.MethodPost)
	router.HandleFunc(fmt.Sprintf("/%s/{%s}", key, key), uh.GetBucketStats).Methods(http.MethodGet)

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
