package store

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
)

func TestOpenUnknownDriver(t *testing.T) {
	_, err := Open(context.Background(), config.Datastore{Driver: "cassandra"})
	if err == nil {
		t.Fatal("expected error for unknown driver")
	}
	if !strings.Contains(err.Error(), "unknown datastore driver") {
		t.Fatalf("err=%v", err)
	}
	if strings.Contains(strings.ToLower(err.Error()), "password") {
		t.Fatalf("error leaked a secret: %v", err)
	}
}

func TestOpenEmptySQLDSN(t *testing.T) {
	_, err := Open(context.Background(), config.Datastore{
		Driver: "postgres",
		SQL:    config.SQLConfig{Driver: "postgres"},
	})
	if err == nil {
		t.Fatal("expected error for empty DSN")
	}
	if !strings.Contains(err.Error(), "dsn") {
		t.Fatalf("err=%v", err)
	}
}

func TestOpenUnsupportedSQLDriver(t *testing.T) {
	_, err := Open(context.Background(), config.Datastore{
		Driver: "sql",
		SQL: config.SQLConfig{
			Driver: "mysql",
			DSN:    "postgres://cabinet:cabinet@localhost:5432/cabinet_studio?sslmode=disable",
		},
	})
	if err == nil {
		t.Fatal("expected error for mysql sql driver")
	}
}

func TestOpenInvalidPostgresDSNNoLeak(t *testing.T) {
	dsn := "postgres://cabinet:super-secret@127.0.0.1:5432/db?sslmode=disable&this=is-not-a-valid-%DSN"
	_, err := Open(context.Background(), config.Datastore{
		Driver: "postgres",
		SQL:    config.SQLConfig{Driver: "postgres", DSN: dsn},
	})
	// May parse successfully (pgx is lenient) or fail. Either way, no secret in error.
	if err != nil && strings.Contains(err.Error(), "super-secret") {
		t.Fatalf("dsn leaked: %v", err)
	}
}

func TestOpenEmptyRedisAddr(t *testing.T) {
	_, err := Open(context.Background(), config.Datastore{Driver: "redis"})
	if err == nil {
		t.Fatal("expected error for empty redis addr")
	}
}

func TestOpenEmptyMongoURI(t *testing.T) {
	_, err := Open(context.Background(), config.Datastore{Driver: "mongodb"})
	if err == nil {
		t.Fatal("expected error for empty mongo uri")
	}
}

func TestOpenRedisConstructsClient(t *testing.T) {
	st, err := Open(context.Background(), config.Datastore{
		Driver: "redis",
		Redis: config.RedisConfig{
			Addr:            "127.0.0.1:6379",
			MaxOpenConns:    0, // must default, not unlimited/zero
			MaxIdleConns:    4,
			ConnMaxLifetime: "1h",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if st.Driver() != "redis" {
		t.Fatalf("driver=%q", st.Driver())
	}
	rs, ok := st.(*redisStore)
	if !ok {
		t.Fatalf("type=%T", st)
	}
	opts := rs.client.Options()
	if opts.PoolSize != 25 {
		t.Fatalf("PoolSize=%d want 25 (zero MaxOpenConns must default)", opts.PoolSize)
	}
	if opts.MinIdleConns != 4 {
		t.Fatalf("MinIdleConns=%d", opts.MinIdleConns)
	}
	if opts.ConnMaxLifetime != time.Hour {
		t.Fatalf("ConnMaxLifetime=%s", opts.ConnMaxLifetime)
	}
}

func TestPoolDefaultingViaConfig(t *testing.T) {
	p := (config.Pool{MaxOpenConns: 0, MaxIdleConns: 10, ConnMaxLifetime: ""}).Normalized()
	if p.MaxOpenConns != 25 {
		t.Fatalf("MaxOpenConns=%d", p.MaxOpenConns)
	}
	if p.MaxIdleConns != 10 {
		t.Fatalf("MaxIdleConns=%d", p.MaxIdleConns)
	}
	if p.ConnMaxLifetime != time.Hour {
		t.Fatalf("ConnMaxLifetime=%s", p.ConnMaxLifetime)
	}
}

func TestNormalizedDriversMapToAdapters(t *testing.T) {
	if (config.Datastore{Driver: "sql"}).NormalizedDriver() != "postgres" {
		t.Fatal("sql")
	}
	if (config.Datastore{Driver: "nosql"}).NormalizedDriver() != "mongodb" {
		t.Fatal("nosql")
	}
}

func TestOpenSQLPoolDefaults(t *testing.T) {
	st, err := Open(context.Background(), config.Datastore{
		Driver: "postgres",
		SQL: config.SQLConfig{
			Driver:          "postgres",
			DSN:             "postgres://cabinet:cabinet@localhost:5432/cabinet_studio?sslmode=disable",
			MaxOpenConns:    0,
			MaxIdleConns:    10,
			ConnMaxLifetime: "1h",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if st.Driver() != "postgres" {
		t.Fatalf("driver=%q", st.Driver())
	}
	ss, ok := st.(*sqlStore)
	if !ok {
		t.Fatalf("type=%T", st)
	}
	if ss.db.Stats().MaxOpenConnections != 25 {
		t.Fatalf("MaxOpenConnections=%d want 25", ss.db.Stats().MaxOpenConnections)
	}
}
