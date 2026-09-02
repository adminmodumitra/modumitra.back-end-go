package outbox

import (
	"context"
	"strings"
	"sync"

	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox/pb"
	"google.golang.org/protobuf/proto"
)

// NoopPublisher succeeds without producing. Local/dev with KAFKA_ENABLED=false
// still drains the table: the worker marks published after this no-op.
type NoopPublisher struct{}

func (NoopPublisher) Publish(context.Context, Record) error { return nil }

// FakePublisher stores Records for tests. Value is protobuf bytes.
type FakePublisher struct {
	mu      sync.Mutex
	Records []Record
	Err     error
}

func (f *FakePublisher) Publish(ctx context.Context, rec Record) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Err != nil {
		return f.Err
	}
	cp := rec
	cp.Value = append([]byte(nil), rec.Value...)
	cp.Headers = append([]Header(nil), rec.Headers...)
	for i := range cp.Headers {
		cp.Headers[i].Value = append([]byte(nil), rec.Headers[i].Value...)
	}
	f.Records = append(f.Records, cp)
	return nil
}

func (f *FakePublisher) Snapshot() []Record {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]Record, len(f.Records))
	copy(out, f.Records)
	return out
}

// UnmarshalEnvelope decodes a Kafka value as Envelope.
func UnmarshalEnvelope(value []byte) (*pb.Envelope, error) {
	var env pb.Envelope
	if err := proto.Unmarshal(value, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// NewPublisher returns a Kafka producer when enabled with brokers,
// otherwise a NoopPublisher so the worker can still drain.
func NewPublisher(enabled bool, brokers []string, clientID string) Publisher {
	var cleaned []string
	for _, b := range brokers {
		b = strings.TrimSpace(b)
		if b != "" {
			cleaned = append(cleaned, b)
		}
	}
	if !enabled || len(cleaned) == 0 {
		return NoopPublisher{}
	}
	if clientID == "" {
		clientID = "cabinet-studio-api"
	}
	return NewKafkaPublisher(cleaned, clientID)
}
