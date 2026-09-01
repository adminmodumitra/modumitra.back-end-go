package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/device"
	"github.com/Danish-13-Khan/backend-cabinet/internal/store"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type Tx struct {
	db *gorm.DB
}

func NewTx(db *gorm.DB) *Tx { return &Tx{db: db} }

func (t *Tx) InTx(ctx context.Context, fn func(context.Context) error) error {
	err := store.GormDB(ctx, t.db).Transaction(func(tx *gorm.DB) error {
		return fn(store.WithGormTx(ctx, tx))
	})
	if err == nil {
		return nil
	}
	var ae *utils.Error
	if errors.As(err, &ae) || errors.Is(err, device.ErrNotFound) {
		return err
	}
	return utils.Generic(err)
}

func use(ctx context.Context, db *gorm.DB) *gorm.DB {
	return store.GormDB(ctx, db)
}

func mapGorm(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return device.ErrNotFound
	}
	var ae *utils.Error
	if errors.As(err, &ae) {
		return ae
	}
	return utils.Generic(err)
}

func utcPtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	u := t.UTC()
	return &u
}

func emptyToNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

var _ device.TxRunner = (*Tx)(nil)
