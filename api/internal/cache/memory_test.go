package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCacheSetGet(t *testing.T) {
	c := NewMemory()
	ctx := context.Background()

	type payload struct{ Name string }

	if err := c.Set(ctx, "key1", payload{"hello"}, time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}

	var got payload
	if err := c.Get(ctx, "key1", &got); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "hello" {
		t.Errorf("Get returned %q, want hello", got.Name)
	}
}

func TestMemoryCacheMissOnAbsentKey(t *testing.T) {
	c := NewMemory()
	var v any
	err := c.Get(context.Background(), "no-such-key", &v)
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss, got %v", err)
	}
}

func TestMemoryCacheTTLExpiry(t *testing.T) {
	c := NewMemory()
	ctx := context.Background()

	_ = c.Set(ctx, "expiring", "value", time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	var v string
	err := c.Get(ctx, "expiring", &v)
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss after TTL, got %v", err)
	}
}

func TestMemoryCacheOverwrite(t *testing.T) {
	c := NewMemory()
	ctx := context.Background()

	_ = c.Set(ctx, "k", "first", time.Minute)
	_ = c.Set(ctx, "k", "second", time.Minute)

	var v string
	_ = c.Get(ctx, "k", &v)
	if v != "second" {
		t.Errorf("expected second write to win, got %q", v)
	}
}
