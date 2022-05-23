package storage

type Bucket interface {
	DecrementKey(key string, entityType string) error
}
