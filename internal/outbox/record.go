package outbox

import (
	"context"
	"errors"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox/pb"
)

var ErrDuplicate = errors.New("outbox event_id already exists")

// Writer inserts a marshaled Envelope into the current GORM transaction.
type Writer interface {
	Enqueue(ctx context.Context, env *pb.Envelope) error
}

// Header is a Kafka record header. Values are raw bytes (not JSON).
type Header struct {
	Key   string
	Value []byte
}

// Record is what Publisher sends. Value is protobuf Envelope bytes only.
type Record struct {
	Topic   string
	Key     string
	Value   []byte
	Headers []Header
}

// Publisher produces one outbox row to Kafka (or a test fake / no-op).
type Publisher interface {
	Publish(ctx context.Context, rec Record) error
}

// Row is a persisted outbox record.
type Row struct {
	PublicID     string
	EventID      string
	Topic        string
	PartitionKey string
	Payload      []byte
	Attempts     int
	LastError    string
	AvailableAt  time.Time
	PublishedAt  *time.Time
	CreatedAt    time.Time
}

// Store is the worker's claim/ack surface (GORM or memory).
type Store interface {
	ClaimDue(ctx context.Context, now time.Time, limit int) ([]Row, error)
	MarkPublished(ctx context.Context, publicID string, at time.Time) error
	MarkFailed(ctx context.Context, publicID string, attempts int, lastErr string, availableAt time.Time) error
}

func headerValue(rec Record, key string) string {
	for _, h := range rec.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}
