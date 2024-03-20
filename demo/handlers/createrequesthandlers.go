package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Azure/rate-limiter/pkg/tokenbucket"
	"github.com/Azure/rate-limiter/ratelimiter"
	"github.com/gorilla/mux"
)

type ClusterCreateRequestHandlers struct {
	ctx         context.Context
	ratelimiter ratelimiter.TokenBucketRateLimiter
	key         string
}

func NewClusterCreateRequestHandlers(ctx context.Context, ratelimiter ratelimiter.TokenBucketRateLimiter, key string) ClusterCreateRequestHandlers {
	return ClusterCreateRequestHandlers{
		ctx:         ctx,
		key:         key,
		ratelimiter: ratelimiter,
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
	log.Printf("find bucket by key: %s\n", id)
	statusCode, err := uh.ratelimiter.GetDecision(id, tokenbucket.DefaultBurstSize, tokenbucket.DefaultTokenDropRatePerMin)
	if statusCode != http.StatusOK {
		// err could only returned for remote cache
		// log and not return error, because we fall back on memcache
		if err != nil {
			log.Printf("failed to get decision from remote cache: %s", err.Error())
		} else if statusCode == http.StatusTooManyRequests {
			http.Error(rw, "too many requests", http.StatusTooManyRequests)
			return
		}
	}

	rw.WriteHeader(http.StatusCreated)
}

func (uh ClusterCreateRequestHandlers) GetBucketStats(rw http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)[uh.key]
	log.Printf("find bucket by key: %s\n", id)
	tokenNumber, err := uh.ratelimiter.GetStats(id, tokenbucket.DefaultBurstSize, tokenbucket.DefaultTokenDropRatePerMin)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(tokenNumber)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		rw.Header().Del("Content-Type")
	}
}
