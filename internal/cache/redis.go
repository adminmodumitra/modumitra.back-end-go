package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Redis backs Cache with go-redis. Errors are logged and treated as misses.
type Redis struct {
	client *redis.Client
}

// NewRedis wraps client. Returns Noop-equivalent behavior only via FromClient.
func NewRedis(client *redis.Client) *Redis {
	return &Redis{client: client}
}

// FromClient returns Redis when client is non-nil, otherwise Noop.
func FromClient(client *redis.Client) Cache {
	if client == nil {
		return Noop{}
	}
	return NewRedis(client)
}

func (r *Redis) Get(ctx context.Context, key string) ([]byte, bool) {
	if r == nil || r.client == nil || key == "" {
		return nil, false
	}
	b, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false
	}
	if err != nil {
		log.Warn().Err(err).Str("key", key).Msg("cache get failed; falling through")
		return nil, false
	}
	return b, true
}

func (r *Redis) Set(ctx context.Context, key string, value []byte, ttl time.Duration) {
	if r == nil || r.client == nil || key == "" {
		return
	}
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		log.Warn().Err(err).Str("key", key).Msg("cache set failed")
	}
}

func (r *Redis) Delete(ctx context.Context, keys ...string) {
	if r == nil || r.client == nil || len(keys) == 0 {
		return
	}
	clean := make([]string, 0, len(keys))
	for _, k := range keys {
		if k != "" {
			clean = append(clean, k)
		}
	}
	if len(clean) == 0 {
		return
	}
	if err := r.client.Del(ctx, clean...).Err(); err != nil {
		log.Warn().Err(err).Strs("keys", clean).Msg("cache delete failed")
	}
}

func (r *Redis) DeletePrefix(ctx context.Context, prefix string) {
	if r == nil || r.client == nil || prefix == "" {
		return
	}
	var cursor uint64
	for {
		keys, next, err := r.client.Scan(ctx, cursor, prefix+"*", 64).Result()
		if err != nil {
			log.Warn().Err(err).Str("prefix", prefix).Msg("cache scan failed")
			return
		}
		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				log.Warn().Err(err).Str("prefix", prefix).Msg("cache prefix delete failed")
			}
		}
		cursor = next
		if cursor == 0 {
			return
		}
	}
}
