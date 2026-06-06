package domain

import "time"

type TransactionLedger struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Quantity  int64     `gorm:"not null"`
	Amount    int64     `gorm:"not null"`
	Status    string    `gorm:"size:255;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (TransactionLedger) TableName() string {
	return "transaction_ledgers"
}
