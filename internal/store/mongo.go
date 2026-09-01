package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"github.com/Danish-13-Khan/backend-cabinet/internal/config"
)

type mongoStore struct {
	client *mongo.Client
	dbName string
}

func openMongo(cfg config.NoSQLConfig) (Store, error) {
	uri := strings.TrimSpace(cfg.URI)
	if uri == "" {
		return nil, fmt.Errorf("mongo uri is required")
	}
	dbName := strings.TrimSpace(cfg.Database)
	if dbName == "" {
		dbName = "cabinet_studio"
	}
	p := cfg.Pool().Normalized()
	opts := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(uint64(p.MaxOpenConns)).
		SetMinPoolSize(uint64(p.MaxIdleConns)).
		SetMaxConnIdleTime(p.ConnMaxLifetime).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(10 * time.Second)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, config.MaskSecrets(fmt.Errorf("mongo connect: %v", err), uri)
	}
	return &mongoStore{client: client, dbName: dbName}, nil
}

func (s *mongoStore) Driver() string { return "mongodb" }

func (s *mongoStore) Ping(ctx context.Context) error {
	return withPingSpan(ctx, "mongodb", func(ctx context.Context) error {
		return s.client.Ping(ctx, readpref.Primary())
	})
}

func (s *mongoStore) Close() error {
	if s == nil || s.client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.client.Disconnect(ctx)
}
