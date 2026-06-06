package domain

import "time"

// PaymentGatewayCommunication stores raw request/response JSON for each gateway call (audit, debugging, reconciliation).
type PaymentGatewayCommunication struct {
	ID                int64     `gorm:"primaryKey;autoIncrement"`
	TransactionID     int64     `gorm:"not null"`
	GatewayName       string    `gorm:"size:255;index;not null"`
	Operation         string    `gorm:"size:255;index;not null"` // e.g. generate_dynamic_qris
	RequestJSON       string    `gorm:"type:jsonb;not null"`
	RequestTimestamp  time.Time `gorm:"not null"`
	ResponseJSON      string    `gorm:"type:jsonb"`
	ResponseStatus    int       `gorm:"index;not null"`
	ResponseTimestamp time.Time `gorm:"not null"`
	PollAttempts      int       `gorm:"not null;default:0"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`
}

func (PaymentGatewayCommunication) TableName() string {
	return "payment_gateway_communications"
}
