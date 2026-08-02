package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type memEntry struct {
	data      []byte
	expiresAt time.Time
}

// MemoryCache is a process-local TTL cache. Safe for concurrent use.
// Entries are lazily evicted on Get; a background ticker sweeps the full map
// every 5 minutes to bound memory growth.
type MemoryCache struct {
	mu sync.RWMutex
	m  map[string]memEntry
}

func NewMemory() Cache {
	c := &MemoryCache{m: make(map[string]memEntry)}
	go c.evictLoop()
	return c
}

func (c *MemoryCache) Get(_ context.Context, key string, v any) error {
	c.mu.RLock()
	e, ok := c.m[key]
	c.mu.RUnlock()

	if !ok || time.Now().After(e.expiresAt) {
		return ErrCacheMiss
	}
	return json.Unmarshal(e.data, v)
}

func (c *MemoryCache) Set(_ context.Context, key string, v any, ttl time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.m[key] = memEntry{data: b, expiresAt: time.Now().Add(ttl)}
	c.mu.Unlock()
	return nil
}

func (c *MemoryCache) evictLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		c.mu.Lock()
		for k, e := range c.m {
			if now.After(e.expiresAt) {
				delete(c.m, k)
			}
		}
		c.mu.Unlock()
	}
}
