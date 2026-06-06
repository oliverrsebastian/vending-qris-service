package repository

import (
	"context"
	"vending-qris-service/internal/domain"

	"gorm.io/gorm"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Save(ctx context.Context, txn *domain.Transaction) (*domain.Transaction, error) {
	if err := r.db.WithContext(ctx).Save(txn).Error; err != nil {
		return nil, err
	}

	return txn, nil
}
