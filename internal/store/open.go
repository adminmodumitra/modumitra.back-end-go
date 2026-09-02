package store

import (
	"context"
	"fmt"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
)

// Set holds every datastore that is enabled at once (Postgres, Redis, Mongo).

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
