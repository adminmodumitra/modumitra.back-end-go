package store

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
)

// Pinger is the readiness dependency. Tests supply a fake.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Closer releases datastore resources.
type Closer interface {
	Close() error
}

// Store is a live adapter (Postgres, Redis, or Mongo) selected by config.
type Store interface {
	Pinger
	Closer
	Driver() string
}

// Open selects and constructs the live store. Unknown drivers fail so the
// process must not boot. Clients are real (not stubs); Ping talks to the server.
func Open(ctx context.Context, cfg config.Datastore) (Store, error) {
	switch cfg.NormalizedDriver() {
	case "postgres":
		return openSQL(ctx, cfg.SQL)
	case "redis":
		return openRedis(cfg.Redis)
	case "mongodb":
		return openMongo(cfg.NoSQL)
	default:
		return nil, fmt.Errorf("unknown datastore driver %q (want postgres, sql, redis, mongodb, or nosql)", strings.TrimSpace(cfg.Driver))
	}
}

func withPingSpan(ctx context.Context, dbSystem string, fn func(context.Context) error) error {
	ctx, span := otel.Tracer("cabinet.store").Start(ctx, "store.ping",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.String("db.system", dbSystem)),
	)
	defer span.End()

	if err := fn(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "ping failed")
		return err
	}
	return nil
}
