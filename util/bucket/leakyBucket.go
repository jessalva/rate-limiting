package bucket

import (
	"context"
	"errors"
	"net/http"
	"sync"
)

type LeakyBucket struct {
	workerPool chan HttpRequestWithResponse
}

func (c *LeakyBucket) WorkerPool() chan HttpRequestWithResponse {
	return c.workerPool
}

func NewLeakyBucket(workerPool chan HttpRequestWithResponse) *LeakyBucket {
	return &LeakyBucket{workerPool: workerPool}
}

type HttpRequestWithResponse struct {
	requestID      string
	request        *http.Request
	ctx            context.Context
	responseWriter http.ResponseWriter
	handler        http.Handler
	wg             *sync.WaitGroup
}

func NewHttpRequestWithResponse(requestID string, request *http.Request, ctx context.Context,
	responseWriter http.ResponseWriter,
	handler http.Handler, wg *sync.WaitGroup) HttpRequestWithResponse {
	return HttpRequestWithResponse{requestID: requestID,
		request: request, ctx: ctx,
		responseWriter: responseWriter,
		handler:        handler, wg: wg}
}

// TryEnqueue  the counter for the given key.
func (c *LeakyBucket) TryEnqueue(job HttpRequestWithResponse) error {

	select {
	case c.workerPool <- job:
		return nil
	default:
		return errors.New("unable to post job to leaky bucket")
	}

}

// TryDequeue 1
func (c *LeakyBucket) TryDequeue() {

	for job := range c.workerPool {
		job.handler.ServeHTTP(job.responseWriter, job.request)
		job.wg.Done()
	}

}
