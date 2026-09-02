package config

import (
	"os"
	"path/filepath"
	"strings"
)

func defaultDatastore() Datastore {
	const poolLife = "1h"
	return Datastore{
		Driver: "postgres",
		SQL: SQLConfig{
			Enabled:         true,
			Driver:          "postgres",
			DSN:             "postgres://cabinet:cabinet@localhost:5432/cabinet_studio?sslmode=disable",
			MaxOpenConns:    defaultMaxOpenConns,
			MaxIdleConns:    10,
			ConnMaxLifetime: poolLife,
		},
		Redis: RedisConfig{
			Addr:            "localhost:6379",
			Password:        "",
			DB:              0,
			TLS:             false,
			MaxOpenConns:    defaultMaxOpenConns,
			MaxIdleConns:    10,
			ConnMaxLifetime: poolLife,
		},
		NoSQL: NoSQLConfig{
			URI:             "mongodb://localhost:27017",
			Database:        "cabinet_studio",
			MaxOpenConns:    defaultMaxOpenConns,
			MaxIdleConns:    10,
			ConnMaxLifetime: poolLife,
		},
	}
}

func resolveDatastorePath() string {
	if p := strings.TrimSpace(os.Getenv("DATASTORE_CONFIG")); p != "" {
		return p
	}
	dir, err := os.Getwd()
	if err != nil {
		return defaultDatastorePath
	}
	for i := 0; i < 8; i++ {
		cand := filepath.Join(dir, defaultDatastorePath)
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return defaultDatastorePath
}
