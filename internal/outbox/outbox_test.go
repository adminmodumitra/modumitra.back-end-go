package outbox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/proto"

	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox/pb"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func traced(t *testing.T) context.Context {
	t.Helper()
	obs.InitTest()
	ctx, span := otel.Tracer("outbox-test").Start(context.Background(), "test")
	t.Cleanup(func() { span.End() })
	return utils.WithRequestID(ctx, "req_testoutbox01")
}

func TestEnqueueMarshalsProtobufNotJSON(t *testing.T) {
	ctx := traced(t)
	w := NewMemoryWriter()
	env := OrgCreated(ctx, "org_abc", "usr_ada", "Lovelace", time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
	if err := w.Enqueue(ctx, env); err != nil {
		t.Fatal(err)
	}
	rows := w.Rows()
	if len(rows) != 1 {
		t.Fatalf("rows=%d", len(rows))
	}
	if bytes.HasPrefix(bytes.TrimSpace(rows[0].Payload), []byte("{")) {
		t.Fatalf("payload looks like JSON: %s", rows[0].Payload)
	}
	var decoded pb.Envelope
	if err := proto.Unmarshal(rows[0].Payload, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.EventType != EventOrgCreated {
		t.Fatalf("event_type=%s", decoded.EventType)
	}
	body := decoded.GetOrgCreated()
	if body == nil || body.OrganizationId != "org_abc" || body.OwnerUserId != "usr_ada" {
		t.Fatalf("body=%+v", body)
	}
	var asJSON map[string]any
	if json.Unmarshal(rows[0].Payload, &asJSON) == nil {
		t.Fatal("payload unmarshaled as JSON object")
	}
	if rows[0].Topic != TopicOrg {
		t.Fatalf("topic=%s", rows[0].Topic)
	}
	if decoded.TraceId == "" {
		t.Fatal("missing trace_id")
	}
	if decoded.RequestId != "req_testoutbox01" {
		t.Fatalf("request_id=%s", decoded.RequestId)
	}
}

func TestWorkerPublishesProtobufAndHeaders(t *testing.T) {
	ctx := traced(t)
	w := NewMemoryWriter()
	env := UserCreated(ctx, "usr_ada", time.Now().UTC())
	if err := w.Enqueue(ctx, env); err != nil {
		t.Fatal(err)
	}
	pub := &FakePublisher{}
	worker := NewWorker(w, pub, SystemClock{})
	if err := worker.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	got := pub.Snapshot()
	if len(got) != 1 {
		t.Fatalf("records=%d", len(got))
	}
	rec := got[0]
	decoded, err := UnmarshalEnvelope(rec.Value)
	if err != nil {
		t.Fatalf("value is not Envelope protobuf: %v", err)
	}
	if decoded.EventType != EventUserCreated {
		t.Fatalf("event_type=%s", decoded.EventType)
	}
	if headerValue(rec, HeaderContentType) != ContentTypeProtobuf {
		t.Fatalf("content-type=%s", headerValue(rec, HeaderContentType))
	}
	if headerValue(rec, HeaderEventID) == "" {
		t.Fatal("missing event_id header")
	}
	if headerValue(rec, HeaderEventType) != EventUserCreated {
		t.Fatalf("event_type header=%s", headerValue(rec, HeaderEventType))
	}
	tp := headerValue(rec, HeaderTraceparent)
	if tp == "" {
		t.Fatal("missing traceparent header")
	}
	if rec.Key != "usr_ada" {
		t.Fatalf("key=%s", rec.Key)
	}
	rows := w.Rows()
	if rows[0].PublishedAt == nil {
		t.Fatal("published_at still null")
	}
}

func TestFailedPublishRetries(t *testing.T) {
	ctx := context.Background()
	clk := &fixedClock{t: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)}
	w := NewMemoryWriter()
	w.now = clk.Now
	if err := w.Enqueue(ctx, UserCreated(ctx, "usr_ada", clk.Now())); err != nil {
		t.Fatal(err)
	}
	pub := &FakePublisher{Err: errors.New("broker down")}
	worker := NewWorker(w, pub, clk)
	if err := worker.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	rows := w.Rows()
	if len(rows) != 1 {
		t.Fatalf("rows=%d", len(rows))
	}
	if rows[0].Attempts != 1 {
		t.Fatalf("attempts=%d", rows[0].Attempts)
	}
	if rows[0].PublishedAt != nil {
		t.Fatal("published_at should still be null")
	}
	if rows[0].LastError == "" {
		t.Fatal("missing last_error")
	}
	if !rows[0].AvailableAt.After(clk.t) {
		t.Fatalf("available_at not backed off: %s", rows[0].AvailableAt)
	}
	if len(pub.Snapshot()) != 0 {
		t.Fatal("failed publish should not store a record")
	}
}

func TestDuplicateEventID(t *testing.T) {
	ctx := context.Background()
	w := NewMemoryWriter()
	env := UserCreated(ctx, "usr_ada", time.Now().UTC())
	env.EventId = "evt_dup_1"
	if err := w.Enqueue(ctx, env); err != nil {
		t.Fatal(err)
	}
	again := UserCreated(ctx, "usr_ada", time.Now().UTC())
	again.EventId = "evt_dup_1"
	if err := w.Enqueue(ctx, again); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("err=%v", err)
	}
	if n := len(w.Rows()); n != 1 {
		t.Fatalf("rows=%d", n)
	}
}

func TestTopicMap(t *testing.T) {
	cases := map[string]string{
		"org.created":                  TopicOrg,
		"membership.created":           TopicOrg,
		"invitation.accepted":          TopicOrg,
		"auth.login":                   TopicAuth,
		"user.created":                 TopicAuth,
		"billing.trial_started":        TopicBilling,
		"billing.subscription_updated": TopicBilling,
		"device.activated":             TopicDevice,
	}
	for ev, want := range cases {
		if got := TopicFor(ev); got != want {
			t.Fatalf("%s -> %s want %s", ev, got, want)
		}
	}
}

func TestNoopPublisherDrains(t *testing.T) {
	ctx := context.Background()
	w := NewMemoryWriter()
	if err := w.Enqueue(ctx, DeviceActivated(ctx, "dev_1", "usr_ada", "macos", time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	worker := NewWorker(w, NoopPublisher{}, SystemClock{})
	if err := worker.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	if w.Rows()[0].PublishedAt == nil {
		t.Fatal("noop should still mark published")
	}
}

type fixedClock struct{ t time.Time }

func (c *fixedClock) Now() time.Time { return c.t }
