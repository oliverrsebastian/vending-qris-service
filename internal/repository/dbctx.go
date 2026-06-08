package repository

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

// WithTx stores an active GORM transaction on ctx for repository calls in the same unit of work.
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// DBFromContext returns the transaction-bound DB when present, otherwise the default connection.
func DBFromContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}
