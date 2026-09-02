package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/sso"
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
	if errors.As(err, &ae) || errors.Is(err, sso.ErrNotFound) || errors.Is(err, sso.ErrInvalidState) || errors.Is(err, sso.ErrDuplicate) {
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

func uniqueConstraint(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName
	}
	return ""
}

func mapGorm(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return sso.ErrNotFound
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

var _ sso.TxRunner = (*Tx)(nil)
