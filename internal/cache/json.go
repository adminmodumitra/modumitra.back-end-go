package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
)

// GetJSON unmarshals a cached value into dest. Returns false on miss or error.
func GetJSON(ctx context.Context, c Cache, key string, dest any) bool {
	if c == nil || dest == nil {
		return false
	}
	b, ok := c.Get(ctx, key)
	if !ok {
		return false
	}
	if err := json.Unmarshal(b, dest); err != nil {
		log.Warn().Err(err).Str("key", key).Msg("cache unmarshal failed; falling through")
		return false
	}
	return true
}

// SetJSON marshals v and stores it. Failures are logged only.
func SetJSON(ctx context.Context, c Cache, key string, v any, ttl time.Duration) {
	if c == nil || v == nil || key == "" {
		return
	}
	b, err := json.Marshal(v)
	if err != nil {
		log.Warn().Err(err).Str("key", key).Msg("cache marshal failed")
		return
	}
	c.Set(ctx, key, b, ttl)
}
