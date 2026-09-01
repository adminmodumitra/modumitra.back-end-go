package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
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
	if errors.As(err, &ae) || errors.Is(err, org.ErrNotFound) {
		return err
	}
	return utils.Generic(err)
}

func use(ctx context.Context, db *gorm.DB) *gorm.DB {
	return store.GormDB(ctx, db)
}

func isUnique(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func mapGorm(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return org.ErrNotFound
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

var _ org.TxRunner = (*Tx)(nil)
