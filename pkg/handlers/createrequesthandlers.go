package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"go.goms.io/token_bucket_cache/tokenbucket"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

type ClusterCreateRequestHandlers struct {
	ctx    context.Context
	client *redis.Client
	key    string
}

func NewClusterCreateRequestHandlers(ctx context.Context, client *redis.Client, key string) ClusterCreateRequestHandlers {
	return ClusterCreateRequestHandlers{
		ctx:    ctx,
		client: client,
		key:    key,
	}
}

func (uh ClusterCreateRequestHandlers) HandleRequest(rw http.ResponseWriter, r *http.Request) {
	payload, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	var u map[string]interface{}
	err = json.Unmarshal([]byte(payload), &u)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	_, found := u[uh.key]
	if !found {
		http.Error(rw, fmt.Sprintf("can't find key %s in request", uh.key), http.StatusInternalServerError)
		return
	}
	id := u[uh.key].(string)
	info, err := uh.client.HGetAll(uh.ctx, id).Result()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	statusCode, err := tokenbucket.UpdateBucketInCache(uh.ctx, uh.client, id, info)
	if err != nil {
		http.Error(rw, err.Error(), statusCode)
		return
	}

	rw.WriteHeader(http.StatusCreated)
}

func (uh ClusterCreateRequestHandlers) GetBucketStats(rw http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)[uh.key]
	fmt.Printf("find bucket by key: %s\n", id)
	info, err := uh.client.HGetAll(r.Context(), id).Result()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(info) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rw.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(info)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		rw.Header().Del("Content-Type")
	}
}
