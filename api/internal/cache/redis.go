package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client     *redis.Client
	defaultTTL time.Duration
}

func NewRedis(rawURL string, defaultTTL time.Duration) (Cache, error) {
	opts, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &RedisCache{client: client, defaultTTL: defaultTTL}, nil
}

func (r *RedisCache) Get(ctx context.Context, key string, v any) error {
	b, err := r.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrCacheMiss
	}
	if err != nil {
		return fmt.Errorf("redis get %q: %w", key, err)
	}
	return json.Unmarshal(b, v)
}

func (r *RedisCache) Set(ctx context.Context, key string, v any, ttl time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal for cache key %q: %w", key, err)
	}
	if ttl == 0 {
		ttl = r.defaultTTL
	}
	return r.client.Set(ctx, key, b, ttl).Err()
}
