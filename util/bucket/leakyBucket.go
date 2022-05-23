package bucket

import (
	"container/list"
	"sync"
)

type LeakyBucket struct {
	bucketLock      sync.Mutex
	processingQueue list.List
	bucketSize      int
}

// Push  the counter for the given key.
func (c *LeakyBucket) Push(key string) {
	c.bucketLock.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.processingQueue.PushFront(key)
	c.bucketLock.Unlock()
}

// Pop returns the current value of the counter for the given key.
func (c *LeakyBucket) Pop(key string) interface{} {
	c.bucketLock.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer c.bucketLock.Unlock()
	frontElement := c.processingQueue.Front()
	c.processingQueue.Remove(frontElement)
	return frontElement.Value
}
