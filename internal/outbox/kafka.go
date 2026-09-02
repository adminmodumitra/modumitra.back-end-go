package outbox

import (
	"context"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaPublisher writes protobuf values to Kafka. Value is never JSON.
type KafkaPublisher struct {
	w *kafka.Writer
}

func NewKafkaPublisher(brokers []string, clientID string) *KafkaPublisher {
	if clientID == "" {
		clientID = "cabinet-studio-api"
	}
	return &KafkaPublisher{
		w: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireAll,
			Async:        false,
			BatchTimeout: 10 * time.Millisecond,
			Transport:    &kafka.Transport{ClientID: clientID},
		},
	}
}

func (p *KafkaPublisher) Publish(ctx context.Context, rec Record) error {
	if p == nil || p.w == nil {
		return nil
	}
	headers := make([]kafka.Header, 0, len(rec.Headers))
	for _, h := range rec.Headers {
		headers = append(headers, kafka.Header{Key: h.Key, Value: h.Value})
	}
	return p.w.WriteMessages(ctx, kafka.Message{
		Topic:   rec.Topic,
		Key:     []byte(rec.Key),
		Value:   rec.Value,
		Headers: headers,
	})
}

func (p *KafkaPublisher) Close() error {
	if p == nil || p.w == nil {
		return nil
	}
	return p.w.Close()
}

func SplitBrokers(raw string) []string {
	var out []string
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
