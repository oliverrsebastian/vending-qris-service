package domain

import "time"

// GatewayPriority is a persisted row defining gateway preference order (lower SortOrder = tried first for new payments).
type GatewayPriority struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	SortOrder   int       `gorm:"not null;uniqueIndex"`
	GatewayName string    `gorm:"size:255;not null;uniqueIndex"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (GatewayPriority) TableName() string {
	return "gateway_priorities"
}
