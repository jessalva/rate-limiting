package middleware

import (
	"RateLimiting/storage"
	"fmt"
	"net/http"
)

type TokenBucketThrottler struct {
	h http.Handler
	s storage.Bucket
}

func NewTokenBucketThrottler(handler http.Handler, s storage.Bucket) *TokenBucketThrottler {
	return &TokenBucketThrottler{h: handler, s: s}
}

func (t *TokenBucketThrottler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	clientName := r.Header.Get("x-MyClientName")
	entityType := r.Header.Get("x-MyEntityType")
	if clientName == "" || entityType == "" {
		clientName = "Browser"
		entityType = "Browser"
		fmt.Println("Client is Browser")
	}

	err := t.s.DecrementKey(clientName, entityType)
	if err != nil {
		fmt.Println(fmt.Sprintf("Got ERROR %s", err))
		w.WriteHeader(http.StatusTooManyRequests)
		_, err = w.Write([]byte("More than 2 request in a minute"))
		if err != nil {
			return
		}
		return
	}

	t.h.ServeHTTP(w, r)

}
