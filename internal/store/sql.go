package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
)

type sqlStore struct {
	db   *sql.DB
	gorm *gorm.DB
}

type gormLogWriter struct{}

func (gormLogWriter) Printf(format string, args ...any) {
	// ParameterizedQueries is on; never interpolate secrets/DSN into logs.
	slog.Warn("gorm", "msg", fmt.Sprintf(format, args...))
}

func newGormLogger() gormlogger.Interface {
	return gormlogger.New(gormLogWriter{}, gormlogger.Config{
		SlowThreshold:             time.Second,
		LogLevel:                  gormlogger.Warn,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      true,
		Colorful:                  false,
	})
}

func openSQL(ctx context.Context, cfg config.SQLConfig) (Store, error) {
	dsn := strings.TrimSpace(cfg.DSN)
	if dsn == "" {
		return nil, fmt.Errorf("sql dsn is required")
	}
	switch strings.ToLower(strings.TrimSpace(cfg.Driver)) {
	case "", "postgres", "postgresql", "pgx":
	default:
		return nil, fmt.Errorf("unsupported sql driver %q (want postgres)", cfg.Driver)
	}

	pgxCfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, config.MaskSecrets(fmt.Errorf("parse postgres dsn: %v", err), dsn)
	}
	pgxCfg.RuntimeParams["timezone"] = "UTC"
	timeout := 10 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining > 0 && remaining < timeout {
			timeout = remaining
		}
	}
	if pgxCfg.ConnectTimeout == 0 || pgxCfg.ConnectTimeout > timeout {
		pgxCfg.ConnectTimeout = timeout
	}

	db := stdlib.OpenDB(*pgxCfg)
	p := cfg.Pool().Normalized()
	// sql.DB SetMaxOpenConns(0) means unlimited — never pass 0.
	db.SetMaxOpenConns(p.MaxOpenConns)
	db.SetMaxIdleConns(p.MaxIdleConns)
	db.SetConnMaxLifetime(p.ConnMaxLifetime)
	db.SetConnMaxIdleTime(p.ConnMaxLifetime)

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{
		NowFunc:              func() time.Time { return time.Now().UTC() },
		TranslateError:       true,
		DisableAutomaticPing: true, // keep Open lazy like database/sql; readiness is store.Ping
		Logger:               newGormLogger(),
	})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open gorm: %w", err)
	}

	return &sqlStore{db: db, gorm: gdb}, nil
}

func (s *sqlStore) Driver() string { return "postgres" }

func (s *sqlStore) DB() *sql.DB { return s.db }

func (s *sqlStore) Gorm() *gorm.DB { return s.gorm }

func (s *sqlStore) Ping(ctx context.Context) error {
	return withPingSpan(ctx, "postgresql", func(ctx context.Context) error {
		return s.db.PingContext(ctx)
	})
}

func (s *sqlStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}
