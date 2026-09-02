package outbox

import "github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"

func rowFrom(m tables.Outbox) Row {
	last := ""
	if m.LastError != nil {
		last = *m.LastError
	}
	payload := m.Payload
	if payload == nil {
		payload = []byte{}
	}
	return Row{
		PublicID:     m.PublicID,
		EventID:      m.EventID,
		Topic:        m.Topic,
		PartitionKey: m.PartitionKey,
		Payload:      append([]byte(nil), payload...),
		Attempts:     m.Attempts,
		LastError:    last,
		AvailableAt:  m.AvailableAt,
		PublishedAt:  m.PublishedAt,
		CreatedAt:    m.CreatedAt,
	}
}
