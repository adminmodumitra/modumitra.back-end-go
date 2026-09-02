package outbox

import "time"

// Model maps schema-v1 outbox. payload is BYTEA protobuf. No AutoMigrate.
type Model struct {
	PublicID     string     `gorm:"column:public_id;primaryKey"`
	EventID      string     `gorm:"column:event_id"`
	Topic        string     `gorm:"column:topic"`
	PartitionKey string     `gorm:"column:partition_key"`
	Payload      []byte     `gorm:"column:payload;type:bytea"`
	Attempts     int        `gorm:"column:attempts"`
	LastError    *string    `gorm:"column:last_error"`
	AvailableAt  time.Time  `gorm:"column:available_at"`
	PublishedAt  *time.Time `gorm:"column:published_at"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime:false"`
}

func (Model) TableName() string { return "outbox" }

func (m Model) toRow() Row {
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
