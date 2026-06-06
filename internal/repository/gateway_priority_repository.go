package repository

import (
	"context"
	"fmt"

	"vending-qris-service/internal/domain"

	"gorm.io/gorm"
)

type GatewayPriorityRepository struct {
	db *gorm.DB
}

func NewGatewayPriorityRepository(db *gorm.DB) *GatewayPriorityRepository {
	return &GatewayPriorityRepository{db: db}
}

func (r *GatewayPriorityRepository) ListOrdered(ctx context.Context) ([]string, error) {
	var rows []domain.GatewayPriority
	if err := r.db.WithContext(ctx).Order("sort_order asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.GatewayName)
	}
	return out, nil
}

func (r *GatewayPriorityRepository) ReplaceAll(ctx context.Context, gateways []string) error {
	if len(gateways) == 0 {
		return fmt.Errorf("gateway priority list must not be empty")
	}

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
		return fmt.Errorf("gateway priority list must contain at least one non-empty gateway id")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM gateway_priorities").Error; err != nil {
			return err
		}
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

// RotateOnce moves the current top-priority gateway to the tail (ops failover without naming a target).
func (r *GatewayPriorityRepository) RotateOnce(ctx context.Context) error {
	names, err := r.ListOrdered(ctx)
	if err != nil {
		return err
	}
	if len(names) < 2 {
		return nil
	}
	rotated := append(append([]string{}, names[1:]...), names[0])
	return r.ReplaceAll(ctx, rotated)
}
