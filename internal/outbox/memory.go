package outbox

import (
	"context"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox/pb"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// MemoryWriter is an in-memory outbox for unit tests (no GORM, no Kafka).
type MemoryWriter struct {
	mu   sync.Mutex
	rows []*Row
	now  func() time.Time
}

func NewMemoryWriter() *MemoryWriter {
	return &MemoryWriter{now: func() time.Time { return time.Now().UTC() }}
}

func (m *MemoryWriter) Enqueue(ctx context.Context, env *pb.Envelope) error {
	if env == nil {
		return nil
	}
	now := m.now()
	id := newPublicID()
	if env.EventId == "" {
		env.EventId = id
	}
	if env.TraceId == "" {
		env.TraceId = obs.TraceIDFromContext(ctx)
	}
	if env.RequestId == "" {
		env.RequestId = utils.RequestIDFromContext(ctx)
	}
	if env.OccurredAt == nil {
		env.OccurredAt = timestamppb.New(now)
	}
	payload, err := proto.Marshal(env)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.rows {
		if r.EventID == env.EventId {
			return ErrDuplicate
		}
	}
	row := &Row{
		PublicID:     id,
		EventID:      env.EventId,
		Topic:        TopicFor(env.EventType),
		PartitionKey: env.PartitionKey,
		Payload:      payload,
		Attempts:     0,
		AvailableAt:  now,
		CreatedAt:    now,
	}
	m.rows = append(m.rows, row)
	return nil
}

func (m *MemoryWriter) Rows() []Row {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Row, 0, len(m.rows))
	for _, r := range m.rows {
		cp := *r
		cp.Payload = append([]byte(nil), r.Payload...)
		if r.PublishedAt != nil {
			t := *r.PublishedAt
			cp.PublishedAt = &t
		}
		out = append(out, cp)
	}
	return out
}

func (m *MemoryWriter) ClaimDue(ctx context.Context, now time.Time, limit int) ([]Row, error) {
	if limit <= 0 {
		limit = DefaultBatch
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Row
	for _, r := range m.rows {
		if r.PublishedAt != nil {
			continue
		}
		if r.AvailableAt.After(now) {
			continue
		}
		cp := *r
		cp.Payload = append([]byte(nil), r.Payload...)
		out = append(out, cp)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (m *MemoryWriter) MarkPublished(ctx context.Context, publicID string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.rows {
		if r.PublicID == publicID {
			t := at.UTC()
			r.PublishedAt = &t
			r.LastError = ""
			return nil
		}
	}
	return nil
}

func (m *MemoryWriter) MarkFailed(ctx context.Context, publicID string, attempts int, lastErr string, availableAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.rows {
		if r.PublicID == publicID {
			r.Attempts = attempts
			r.LastError = lastErr
			r.AvailableAt = availableAt.UTC()
			return nil
		}
	}
	return nil
}

var (
	_ Writer = (*MemoryWriter)(nil)
	_ Store  = (*MemoryWriter)(nil)
)
