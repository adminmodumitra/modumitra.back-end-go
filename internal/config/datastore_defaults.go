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

func resolveConfigPath(envKey, defaultRel string) string {
	if p := strings.TrimSpace(os.Getenv(envKey)); p != "" {
		return p
	}
	return findRel(defaultRel)
}

func resolveDatastorePath() string {
	return resolveConfigPath("DATASTORE_CONFIG", defaultDatastorePath)
}

func findRel(rel string) string {
	dir, err := os.Getwd()
	if err != nil {
		return rel
	}
	for i := 0; i < 8; i++ {
		cand := filepath.Join(dir, rel)
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return rel
}
