package handlers

import "net/http"

type RequestHandler interface {
	HandleRequest(rw http.ResponseWriter, r *http.Request)
	GetBucketStats(rw http.ResponseWriter, r *http.Request)
}
