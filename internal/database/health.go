package database

import (
	"context"

	"gorm.io/gorm"
)

// Health implements controller.HealthChecker using the underlying SQL connection.
type Health struct {
	db *gorm.DB
}

func NewHealth(db *gorm.DB) *Health {
	return &Health{db: db}
}

func (h *Health) Ping(ctx context.Context) error {
	sqlDB, err := h.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
