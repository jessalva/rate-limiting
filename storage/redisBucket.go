package storage

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis"
	"log"
	"time"
)

type redisBucket struct {
	client  *redis.Client
	redSync *redsync.Redsync
}

func NewRedisBucket(client *redis.Client) Bucket {

	fmt.Println("Testing Golang Redis")
	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	return &redisBucket{client: client, redSync: rs}
}

func (r *redisBucket) DecrementKey(key string, entityType string) error {

	redisKey := fmt.Sprintf("%s.%s", entityType, key)
	redisLockForKey := fmt.Sprintf("%s.LOCK", redisKey)

	currTokenCount, err := r.client.Get(redisKey).Int()
	if err == redis.Nil {
		currTokenCount = 2
		log.Println("key does not exist")
	} else if err != nil {
		log.Println("Get failed", err)
		return err
	} else if currTokenCount == 0 {
		return errors.New("all tokens are consumed")
	} else {
		log.Println(fmt.Sprintf("Number of token for %s: %d", redisKey, currTokenCount))
	}

	mutex := r.redSync.NewMutex(redisLockForKey)

	if err := mutex.Lock(); err != nil {
		return err
	}
	defer func(mutex *redsync.Mutex) {
		_, err := mutex.Unlock()
		if err != nil {
			log.Println("Unlocking failed", mutex.Name(), err)
		}
	}(mutex)

	r.client.Set(redisKey, currTokenCount-1, 10*time.Second)

	return nil
}
