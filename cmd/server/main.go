package main

import (
	"RateLimiting/middleware"
	"RateLimiting/storage"
	"RateLimiting/util/bucket"
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
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

	requestWorkerChannel := make(chan bucket.HttpRequestWithResponse, 2)
	leakyBucket := bucket.NewLeakyBucket(requestWorkerChannel)
	defer close(leakyBucket.WorkerPool())
	go leakyBucket.TryDequeue()

	wrappedMux := middleware.NewLeakyBucketThrottler(middleware.NewTokenBucketThrottler(mux, redisBucket), leakyBucket)

	server := &http.Server{Addr: ":9000", Handler: wrappedMux}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			// handle err
			log.Fatalf("got fatal error while starting server: %s", err)
		}
	}()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, os.Kill)

	// Waiting for SIGINT (kill -2)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("got error while shutting down server: %s", err)
	}

}
