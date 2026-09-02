package outbox

import (
	"context"
	"errors"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"

	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox/pb"
	"github.com/Danish-13-Khan/backend-cabinet/internal/store"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

// SQLWriter inserts outbox rows via GORM into the current transaction.
type SQLWriter struct {
	db  *gorm.DB
	now func() time.Time
}

func NewWriter(db *gorm.DB) *SQLWriter {
	return &SQLWriter{db: db, now: func() time.Time { return time.Now().UTC() }}
}

func (w *SQLWriter) Enqueue(ctx context.Context, env *pb.Envelope) error {
	if w == nil || w.db == nil {
		return errors.New("outbox writer: no database")
	}
	if env == nil {
		return errors.New("outbox writer: nil envelope")
	}
	now := w.now()
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
	row := tables.Outbox{
		PublicID:     id,
		EventID:      env.EventId,
		Topic:        TopicFor(env.EventType),
		PartitionKey: env.PartitionKey,
		Payload:      payload,
		Attempts:     0,
		AvailableAt:  now,
		CreatedAt:    now,
	}
	err = store.GormDB(ctx, w.db).Create(&row).Error
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrDuplicate
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrDuplicate
	}
	return err
}

var _ Writer = (*SQLWriter)(nil)
