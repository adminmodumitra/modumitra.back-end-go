package cache

import (
	"context"
	"time"
)

// Cache is a best-effort byte store. Misses and errors both return ok=false;
// callers must fall through to the source of truth.
type Cache interface {
	Get(ctx context.Context, key string) (value []byte, ok bool)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration)
	Delete(ctx context.Context, keys ...string)
	// DeletePrefix removes keys with the given prefix (SCAN). Best-effort.
	DeletePrefix(ctx context.Context, prefix string)
}

// Noop is used when Redis is not in EnabledDrivers.
type Noop struct{}

func (Noop) Get(context.Context, string) ([]byte, bool)         { return nil, false }
func (Noop) Set(context.Context, string, []byte, time.Duration) {}
func (Noop) Delete(context.Context, ...string)                  {}
func (Noop) DeletePrefix(context.Context, string)               {}

// New returns c, or Noop when c is nil.
func New(c Cache) Cache {
	if c == nil {
		return Noop{}
	}
	return c
}
