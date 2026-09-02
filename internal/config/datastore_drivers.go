package config

import (
	"strings"
	"time"
)

const (
	defaultDatastorePath = "configs/datastore.dummy.json"
	defaultMaxOpenConns  = 25
	defaultConnLifetime  = time.Hour
)

type Datastore struct {
	Driver  string      `json:"Driver"`  // single-driver fallback: "postgres" | "sql" | "redis" | "mongodb" | "nosql"
	Drivers []string    `json:"Drivers"` // concurrent adapters, e.g. ["postgres","redis","mongodb"]
	SQL     SQLConfig   `json:"SQL"`
	Redis   RedisConfig `json:"Redis"`
	NoSQL   NoSQLConfig `json:"NoSQL"`
}

// EnabledDrivers is the set of adapters to open at once.
// Priority: Drivers list (or DATASTORE_DRIVERS) → per-block Enabled → single Driver fallback.
func (d Datastore) EnabledDrivers() []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(raw string) {
		n := normalizeDriverName(raw)
		if n == "" {
			return
		}
		if n != "postgres" && n != "redis" && n != "mongodb" {
			n = strings.ToLower(strings.TrimSpace(raw))
		}
		if _, ok := seen[n]; ok {
			return
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	for _, raw := range d.Drivers {
		add(raw)
	}
	if len(out) > 0 {
		return out
	}
	if d.SQL.Enabled {
		add("postgres")
	}
	if d.Redis.Enabled {
		add("redis")
	}
	if d.NoSQL.Enabled {
		add("mongodb")
	}
	if len(out) > 0 {
		return out
	}
	add(d.NormalizedDriver())
	return out
}

func normalizeDriverName(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "postgres", "postgresql", "sql", "pgx":
		if strings.TrimSpace(s) == "" {
			return ""
		}
		return "postgres"
	case "redis":
		return "redis"
	case "mongodb", "mongo", "nosql":
		return "mongodb"
	default:
		return strings.ToLower(strings.TrimSpace(s))
	}
}

// NormalizedDriver returns the adapter key: postgres, redis, or mongodb.
// Empty driver is Postgres (v1 product default). Unknown values are returned lowercased
// so Open can reject them.
func (d Datastore) NormalizedDriver() string {
	switch strings.ToLower(strings.TrimSpace(d.Driver)) {
	case "", "postgres", "postgresql", "sql", "pgx":
		return "postgres"
	case "redis":
		return "redis"
	case "mongodb", "mongo", "nosql":
		return "mongodb"
	default:
		return strings.ToLower(strings.TrimSpace(d.Driver))
	}
}

// ReadyCheckName is the /readyz checks key for the active driver.
func (d Datastore) ReadyCheckName() string {
	switch d.NormalizedDriver() {
	case "redis":
		return "redis"
	case "mongodb":
		return "mongodb"
	default:
		return "postgres"
	}
}
