package store

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"github.com/redis/go-redis/v9"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
)

type redisStore struct {
	client *redis.Client
}

func openRedis(cfg config.RedisConfig) (Store, error) {
	addr := strings.TrimSpace(cfg.Addr)
	if addr == "" {
		return nil, fmt.Errorf("redis addr is required")
	}
	p := cfg.Pool().Normalized()
	opts := &redis.Options{
		Addr:            addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        p.MaxOpenConns,
		MinIdleConns:    p.MaxIdleConns,
		ConnMaxLifetime: p.ConnMaxLifetime,
	}
	if cfg.TLS {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr
		}
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: host,
		}
	}
	return &redisStore{client: redis.NewClient(opts)}, nil
}

func (s *redisStore) Driver() string { return "redis" }

func (s *redisStore) Client() *redis.Client { return s.client }

func (s *redisStore) Ping(ctx context.Context) error {
	return withPingSpan(ctx, "redis", func(ctx context.Context) error {
		return s.client.Ping(ctx).Err()
	})
}

func (s *redisStore) Close() error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Close()
}
