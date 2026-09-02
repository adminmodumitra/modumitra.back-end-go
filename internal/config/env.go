package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func applyDatastoreEnv(ds *Datastore) {
	if v := strings.TrimSpace(os.Getenv("DATASTORE_DRIVER")); v != "" {
		ds.Driver = v
	}
	if v := strings.TrimSpace(os.Getenv("DATASTORE_DRIVERS")); v != "" {
		var parts []string
		for _, p := range strings.Split(v, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				parts = append(parts, p)
			}
		}
		if len(parts) > 0 {
			ds.Drivers = parts
		}
	}
	if v := strings.TrimSpace(os.Getenv("DATABASE_URL")); v != "" {
		ds.SQL.DSN = v
	}
	if v := strings.TrimSpace(os.Getenv("REDIS_ADDR")); v != "" {
		ds.Redis.Addr = v
	}
	if v := strings.TrimSpace(os.Getenv("REDIS_PASSWORD")); v != "" {
		ds.Redis.Password = v
	}
	if v := strings.TrimSpace(os.Getenv("REDIS_DB")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			ds.Redis.DB = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("REDIS_TLS")); v != "" {
		ds.Redis.TLS = getenvBool("REDIS_TLS", ds.Redis.TLS)
	}
	if v := strings.TrimSpace(os.Getenv("MONGO_URI")); v != "" {
		ds.NoSQL.URI = v
	}
	if v := strings.TrimSpace(os.Getenv("MONGO_DATABASE")); v != "" {
		ds.NoSQL.Database = v
	}
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getenvInt64(key string, fallback int64) int64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}
	return n
}

func getenvBool(key string, fallback bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

func splitCSV(raw string) []string {
	var out []string
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
