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

	"pkg/azureclient"
	"pkg/handlers"

	"github.com/gorilla/mux"
)

const (
	key                  = "billingAccount"
	bucketMaxTokenNumber = 10
	tokenDropRatePerMin  = 1
)

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

	redisClient, err := azureclient.BuildRedisClientFromAzure(ctx, subscriptionID, resourceGroupName, redisName)
	if err != nil {
		log.Fatal(err)
	}

	uh := handlers.NewClusterCreateRequestHandlers(ctx, redisClient, key)

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
