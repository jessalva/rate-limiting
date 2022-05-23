package main

import (
	"RateLimiting/middleware"
	"RateLimiting/storage"
	"fmt"
	"github.com/go-redis/redis"
	"net/http"
)

func handleRequest(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Call succeeded")
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRequest)

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	redisBucket := storage.NewRedisBucket(client)

	wrappedMux := middleware.NewTokenBucketThrottler(mux, redisBucket)

	http.ListenAndServe(":8000", wrappedMux)

}
