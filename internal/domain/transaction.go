package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID            int64           `gorm:"primaryKey;autoIncrement"`
	InvoiceNumber string          `gorm:"size:255;index"`
	Products      []Product       `gorm:"name:products;type:jsonb;not null"`
	Amount        decimal.Decimal `gorm:"not null"`
	Status        string          `gorm:"size:255;not null"`
	CreatedAt     time.Time       `gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime"`
}

type Product struct {
	Name      string          `json:"name"`
	Quantity  int             `json:"quantity"`
	ItemPrice decimal.Decimal `json:"item_price"`
}

func (*Transaction) TableName() string {
	return "transactions"
}

func (t *Transaction) IsPaid() bool {
	return t.Status == "PAID"
}

func (t *Transaction) UpdateToPaid() {
	t.Status = "PAID"
}
