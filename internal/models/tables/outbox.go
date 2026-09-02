package tables

import "time"

// Outbox maps schema-v1 outbox. payload is BYTEA protobuf. No AutoMigrate.
type Outbox struct {
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

func (Outbox) TableName() string { return "outbox" }
