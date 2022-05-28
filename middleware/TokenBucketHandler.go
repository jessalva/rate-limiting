package middleware

import (
	"RateLimiting/storage"
	"fmt"
	"log"
	"net/http"
)

type TokenBucketHandler struct {
	h http.Handler
	s storage.Bucket
}

func NewTokenBucketThrottler(handler http.Handler, s storage.Bucket) *TokenBucketHandler {
	return &TokenBucketHandler{h: handler, s: s}
}

func (t *TokenBucketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	clientName := r.Header.Get("x-MyClientName")
	entityType := r.Header.Get("x-MyEntityType")
	if clientName == "" || entityType == "" {
		clientName = "Browser"
		entityType = "Browser"
		log.Println("Client is Browser")
	}

	err := t.s.DecrementKey(clientName, entityType)
	if err != nil {
		log.Println(fmt.Sprintf("Got ERROR %s", err))
		w.WriteHeader(http.StatusTooManyRequests)
		_, err = w.Write([]byte("More than 2 request in a minute by TokenBucketHandler"))
		if err != nil {
			return
		}
		return
	}

	log.Println("Reached here")

	t.h.ServeHTTP(w, r)

}
