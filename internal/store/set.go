package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
)

// Set holds every datastore that is enabled at once (Postgres, Redis, Mongo).
type Set struct {
	items []Store
}

func (s *Set) Driver() string {
	if s == nil || len(s.items) == 0 {
		return ""
	}
	if len(s.items) == 1 {
		return s.items[0].Driver()
	}
	names := make([]string, 0, len(s.items))
	for _, it := range s.items {
		names = append(names, it.Driver())
	}
	return strings.Join(names, ",")
}

func (s *Set) Ping(ctx context.Context) error {
	_, err := s.Checks(ctx)
	return err
}

// Checks pings each live adapter. The map always includes every driver;
// the returned error is the first failure (so /readyz can 503).
func (s *Set) Checks(ctx context.Context) (map[string]string, error) {
	out := map[string]string{}
	if s == nil {
		return out, fmt.Errorf("no datastore")
	}
	var first error
	for _, it := range s.items {
		name := it.Driver()
		if err := it.Ping(ctx); err != nil {
			out[name] = "error"
			if first == nil {
				first = fmt.Errorf("%s: %w", name, err)
			}
			continue
		}
		out[name] = "ok"
	}
	return out, first
}

func (s *Set) Close() error {
	if s == nil {
		return nil
	}
	var first error
	for _, it := range s.items {
		if err := it.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (s *Set) SQL() Store { return s.find("postgres") }

// SQLDB returns the live Postgres *sql.DB, or nil if Postgres is not enabled.
func (s *Set) SQLDB() *sql.DB {
	st := s.SQL()
	if st == nil {
		return nil
	}
	type dber interface{ DB() *sql.DB }
	d, ok := st.(dber)
	if !ok {
		return nil
	}
	return d.DB()
}
func (s *Set) Redis() Store { return s.find("redis") }
func (s *Set) Mongo() Store { return s.find("mongodb") }

func (s *Set) find(name string) Store {
	if s == nil {
		return nil
	}
	for _, it := range s.items {
		if it.Driver() == name {
			return it
		}
	}
	return nil
}

func (s *Set) Len() int {
	if s == nil {
		return 0
	}
	return len(s.items)
}

// OpenAll connects every enabled adapter. One failure closes those already
// opened and returns. Unknown names fail boot.
func OpenAll(ctx context.Context, cfg config.Datastore) (*Set, error) {
	drivers := cfg.EnabledDrivers()
	if len(drivers) == 0 {
		return nil, fmt.Errorf("no datastore enabled")
	}
	set := &Set{}
	for _, d := range drivers {
		st, err := openOne(ctx, d, cfg)
		if err != nil {
			_ = set.Close()
			return nil, err
		}
		set.items = append(set.items, st)
	}
	return set, nil
}

func openOne(ctx context.Context, driver string, cfg config.Datastore) (Store, error) {
	switch driver {
	case "postgres":
		return openSQL(ctx, cfg.SQL)
	case "redis":
		return openRedis(cfg.Redis)
	case "mongodb":
		return openMongo(cfg.NoSQL)
	default:
		return nil, fmt.Errorf("unknown datastore driver %q (want postgres, sql, redis, mongodb, or nosql)", driver)
	}
}
