package middleware

import (
	"RateLimiting/util/bucket"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"sync"
	"time"
)

type LeakyBucketHandler struct {
	h http.Handler
	s *bucket.LeakyBucket
}

func NewLeakyBucketThrottler(handler http.Handler, s *bucket.LeakyBucket) *LeakyBucketHandler {
	return &LeakyBucketHandler{h: handler, s: s}
}

func (t *LeakyBucketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	clientName := r.Header.Get("x-MyClientName")
	entityType := r.Header.Get("x-MyEntityType")
	reqId := r.Header.Get("x-RequestID")
	if clientName == "" || entityType == "" {
		clientName = "Browser"
		entityType = "Browser"
	}
	if reqId == "" {
		uuidReqID, _ := uuid.NewUUID()
		reqId = uuidReqID.String()
		log.Println(fmt.Sprintf("RequestId: %s", reqId))
	}

	var wg sync.WaitGroup
	wg.Add(1)
	reqWithResponse := bucket.NewHttpRequestWithResponse(reqId, r, r.Context(), w, t.h, &wg)

	err := t.s.TryEnqueue(reqWithResponse)
	if err != nil {
		log.Println(fmt.Sprintf("Got ERROR %s", err))
		w.WriteHeader(http.StatusTooManyRequests)
		_, err = w.Write([]byte("More than 3 request in a minute by LeakyBucketHandler"))
		if err != nil {
			log.Println(fmt.Sprintf("Got ERROR while writing %s", err))
			return
		}
		return
	}

	if waitTimedOut(&wg, time.Second) {
		log.Println(fmt.Sprintf("Got Timed out"))
		w.WriteHeader(http.StatusRequestTimeout)
		_, err = w.Write([]byte("Request timed out"))
		if err != nil {
			log.Println(fmt.Sprintf("Got ERROR while writing %s", err))
			return
		}
		return
	}

}

func waitTimedOut(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		wg.Done()
		return true // timed out
	}
}
