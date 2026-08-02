package cache

import (
	"context"
	"time"
)

// Noop is a no-op cache that always misses. Used when Redis is unavailable.
type Noop struct{}

func NewNoop() Cache { return &Noop{} }

func (n *Noop) Get(_ context.Context, _ string, _ any) error {
	return ErrCacheMiss
}
func (n *Noop) Set(_ context.Context, _ string, _ any, _ time.Duration) error { return nil }
