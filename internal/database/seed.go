package database

import (
	"context"
	"fmt"

	"vending-qris-service/internal/domain"

	"gorm.io/gorm"
)

// SeedGatewayPriorities inserts the default gateway priority list when the table is empty.
func SeedGatewayPriorities(ctx context.Context, db *gorm.DB, bootstrap []string) error {
	var count int64
	if err := db.WithContext(ctx).Model(&domain.GatewayPriority{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if len(bootstrap) == 0 {
		return fmt.Errorf("cannot seed gateway_priorities: empty bootstrap")
	}
	return seedGatewayPriorities(ctx, db, bootstrap)
}

func seedGatewayPriorities(ctx context.Context, db *gorm.DB, gateways []string) error {
	seen := make(map[string]struct{})
	dedup := make([]string, 0, len(gateways))
	for _, g := range gateways {
		if g == "" {
			continue
		}
		if _, ok := seen[g]; ok {
			continue
		}
		seen[g] = struct{}{}
		dedup = append(dedup, g)
	}
	if len(dedup) == 0 {
		return fmt.Errorf("gateway priority seed must contain at least one non-empty gateway id")
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, name := range dedup {
			row := domain.GatewayPriority{
				SortOrder:   i,
				GatewayName: name,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
