package outbox

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/protobuf/proto"

	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox/pb"
)

const (
	DefaultBatch    = 32
	DefaultInterval = 500 * time.Millisecond
)

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

// Worker claims unpublished rows, produces protobuf records, and acks.
// At-least-once: consumers must be idempotent on event_id.
type Worker struct {
	store    Store
	pub      Publisher
	clock    Clock
	batch    int
	interval time.Duration
}

func NewWorker(store Store, pub Publisher, clock Clock) *Worker {
	if clock == nil {
		clock = SystemClock{}
	}
	if pub == nil {
		pub = NoopPublisher{}
	}
	return &Worker{
		store:    store,
		pub:      pub,
		clock:    clock,
		batch:    DefaultBatch,
		interval: DefaultInterval,
	}
}

func (w *Worker) Run(ctx context.Context) {
	if w == nil || w.store == nil {
		return
	}
	t := time.NewTicker(w.interval)
	defer t.Stop()
	for {
		if err := w.Tick(ctx); err != nil {
			l := obs.FromCtx(ctx)
			l.Error().Err(err).Msg("outbox.tick")
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (w *Worker) Tick(ctx context.Context) error {
	if w == nil || w.store == nil {
		return nil
	}
	now := w.clock.Now()
	rows, err := w.store.ClaimDue(ctx, now, w.batch)
	if err != nil {
		return err
	}
	for i := range rows {
		if err := ctx.Err(); err != nil {
			return err
		}
		if pubErr := w.publish(ctx, rows[i]); pubErr != nil {
			attempts := rows[i].Attempts + 1
			avail := now.Add(backoff(attempts))
			if markErr := w.store.MarkFailed(ctx, rows[i].PublicID, attempts, pubErr.Error(), avail); markErr != nil {
				l := obs.FromCtx(ctx)
				l.Error().Err(markErr).Str("event_id", rows[i].EventID).Msg("outbox.mark_failed")
			}
			continue
		}
		if markErr := w.store.MarkPublished(ctx, rows[i].PublicID, now); markErr != nil {
			l := obs.FromCtx(ctx)
			l.Error().Err(markErr).Str("event_id", rows[i].EventID).Msg("outbox.mark_published")
		}
	}
	return nil
}

func (w *Worker) publish(ctx context.Context, row Row) error {
	headers := []Header{
		{Key: HeaderContentType, Value: []byte(ContentTypeProtobuf)},
		{Key: HeaderEventID, Value: []byte(row.EventID)},
	}
	var env pb.Envelope
	if err := proto.Unmarshal(row.Payload, &env); err == nil {
		headers = append(headers, Header{Key: HeaderEventType, Value: []byte(env.EventType)})
	} else {
		headers = append(headers, Header{Key: HeaderEventType, Value: []byte(row.Topic)})
	}
	if tp := injectTraceparent(ctx); tp != "" {
		headers = append(headers, Header{Key: HeaderTraceparent, Value: []byte(tp)})
	}
	return w.pub.Publish(ctx, Record{
		Topic:   row.Topic,
		Key:     row.PartitionKey,
		Value:   row.Payload,
		Headers: headers,
	})
}

func injectTraceparent(ctx context.Context) string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier.Get("traceparent")
}

func backoff(attempts int) time.Duration {
	switch {
	case attempts <= 1:
		return time.Second
	case attempts == 2:
		return 5 * time.Second
	default:
		return 30 * time.Second
	}
}
