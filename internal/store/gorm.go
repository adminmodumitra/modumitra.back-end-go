package store

import (
	"context"

	"gorm.io/gorm"
)

type gormTxKey struct{}

// WithGormTx stores a GORM transaction on ctx so nested repo calls share it.
func WithGormTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, gormTxKey{}, tx)
}

// GormDB returns the in-flight transaction if present, otherwise fallback
// bound to ctx. Always call this instead of using *gorm.DB directly.
func GormDB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(gormTxKey{}).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx)
	}
	return fallback.WithContext(ctx)
}
