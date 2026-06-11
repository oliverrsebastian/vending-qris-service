package repository

import (
	"context"
	"errors"
	"fmt"

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
	if err := DBFromContext(ctx, r.db).Save(txn).Error; err != nil {
		return nil, err
	}

	return txn, nil
}

func (r *TransactionRepository) FindByInvoiceNumber(ctx context.Context, invoiceNumber string) (*domain.Transaction, error) {
	var txn domain.Transaction
	err := DBFromContext(ctx, r.db).Where("invoice_number = ?", invoiceNumber).First(&txn).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &txn, nil
}

func (r *TransactionRepository) FindByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	var txn domain.Transaction
	err := DBFromContext(ctx, r.db).First(&txn, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("transaction %d not found", id)
	}
	if err != nil {
		return nil, err
	}
	return &txn, nil
}
