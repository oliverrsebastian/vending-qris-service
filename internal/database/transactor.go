package database

import (
	"context"

	"vending-qris-service/internal/domain"
	"vending-qris-service/internal/repository"

	"gorm.io/gorm"
)

type transactor struct {
	db *gorm.DB
}

func NewTransactor(db *gorm.DB) domain.Transactor {
	return &transactor{db: db}
}

func (t *transactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(repository.WithTx(ctx, tx))
	})
}
