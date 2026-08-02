package cache

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss is returned by Get when the key is not in the cache.
var ErrCacheMiss = errors.New("cache miss")

// Cache is a read-through/write-through store for arbitrary JSON-serialisable values.
type Cache interface {
	Get(ctx context.Context, key string, v any) error
	Set(ctx context.Context, key string, v any, ttl time.Duration) error
}
