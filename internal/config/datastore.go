package config

import (
	"strings"
	"time"
)

// Pool is the shared connection-pool shape. SQL, Redis, and NoSQL repeat these fields.
type Pool struct {
	MaxOpenConns    int    `json:"MaxOpenConns"`
	MaxIdleConns    int    `json:"MaxIdleConns"`
	ConnMaxLifetime string `json:"ConnMaxLifetime"` // Go duration, e.g. "1h"
}

// NormalizedPool is Pool after defaults (never pass 0 as sql MaxOpenConns).
type NormalizedPool struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func (p Pool) Normalized() NormalizedPool {
	maxOpen := p.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = defaultMaxOpenConns
	}
	maxIdle := p.MaxIdleConns
	if maxIdle < 0 {
		maxIdle = 0
	}
	if maxIdle > maxOpen {
		maxIdle = maxOpen
	}
	life := defaultConnLifetime
	if s := strings.TrimSpace(p.ConnMaxLifetime); s != "" {
		if d, err := time.ParseDuration(s); err == nil && d > 0 {
			life = d
		}
	}
	return NormalizedPool{
		MaxOpenConns:    maxOpen,
		MaxIdleConns:    maxIdle,
		ConnMaxLifetime: life,
	}
}

type SQLConfig struct {
	Enabled         bool   `json:"Enabled"`
	Driver          string `json:"Driver"` // "postgres" (required for sql)
	DSN             string `json:"DSN"`    // postgres://... local or remote
	MaxOpenConns    int    `json:"MaxOpenConns"`
	MaxIdleConns    int    `json:"MaxIdleConns"`
	ConnMaxLifetime string `json:"ConnMaxLifetime"`
}

func (c SQLConfig) Pool() Pool {
	return Pool{
		MaxOpenConns:    c.MaxOpenConns,
		MaxIdleConns:    c.MaxIdleConns,
		ConnMaxLifetime: c.ConnMaxLifetime,
	}
}

type RedisConfig struct {
	Enabled         bool   `json:"Enabled"`
	Addr            string `json:"Addr"` // host:port
	Password        string `json:"Password"`
	DB              int    `json:"DB"`
	TLS             bool   `json:"TLS"`
	MaxOpenConns    int    `json:"MaxOpenConns"`    // maps to go-redis PoolSize
	MaxIdleConns    int    `json:"MaxIdleConns"`    // MinIdleConns
	ConnMaxLifetime string `json:"ConnMaxLifetime"` // ConnMaxLifetime
}

func (c RedisConfig) Pool() Pool {
	return Pool{
		MaxOpenConns:    c.MaxOpenConns,
		MaxIdleConns:    c.MaxIdleConns,
		ConnMaxLifetime: c.ConnMaxLifetime,
	}
}

type NoSQLConfig struct {
	Enabled         bool   `json:"Enabled"`
	URI             string `json:"URI"` // mongodb:// or mongodb+srv://
	Database        string `json:"Database"`
	MaxOpenConns    int    `json:"MaxOpenConns"`    // MaxPoolSize
	MaxIdleConns    int    `json:"MaxIdleConns"`    // MinPoolSize
	ConnMaxLifetime string `json:"ConnMaxLifetime"` // MaxConnIdleTime
}

func (c NoSQLConfig) Pool() Pool {
	return Pool{
		MaxOpenConns:    c.MaxOpenConns,
		MaxIdleConns:    c.MaxIdleConns,
		ConnMaxLifetime: c.ConnMaxLifetime,
	}
}
