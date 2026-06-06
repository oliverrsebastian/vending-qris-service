package domain

import "time"

// PaymentGatewayCommunication stores raw request/response JSON for each gateway call (audit, debugging, reconciliation).
type PaymentGatewayCommunication struct {
	ID                  int64     `gorm:"primaryKey;autoIncrement"`
	TransactionLedgerID int64     `gorm:"not null"`
	GatewayName         string    `gorm:"size:255;index;not null"`
	Operation           string    `gorm:"size:255;index;not null"` // e.g. generate_dynamic_qris
	RequestJSON         []byte    `gorm:"type:jsonb;not null"`
	RequestTimestamp    time.Time `gorm:"not null"`
	ResponseJSON        []byte    `gorm:"type:jsonb"`
	ResponseStatus      string    `gorm:"size:255;index;not null"`
	ResponseTimestamp   time.Time `gorm:"not null"`
	PollAttempts        int       `gorm:"not null;default:0"`
	CreatedAt           time.Time `gorm:"autoCreateTime"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime"`
}

func (PaymentGatewayCommunication) TableName() string {
	return "payment_gateway_communications"
}
