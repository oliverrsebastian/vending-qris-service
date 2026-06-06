package gateway

import (
	"context"
	"fmt"

	"vending-qris-service/internal/domain"
)

// Resolver picks the first gateway in the priority list that constructs and Ping succeeds.
type Resolver struct {
	priority domain.GatewayPriorityRepository
}

func NewResolver(priority domain.GatewayPriorityRepository) *Resolver {
	return &Resolver{priority: priority}
}

func (r *Resolver) Resolve(ctx context.Context) (domain.PaymentGateway, error) {
	if r == nil || r.priority == nil {
		return nil, fmt.Errorf("gateway resolver not configured")
	}

	names, err := r.priority.ListOrdered(ctx)
	if err != nil {
		return nil, err
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("gateway priority list is empty; PUT /v1/admin/payment-gateways/priority")
	}
	var lastPingErr error
	for _, name := range names {
		gw, err := New(name)
		if err != nil {
			lastPingErr = err
			continue
		}

		if err := gw.Ping(ctx); err != nil {
			lastPingErr = err
			continue
		}

		return gw, nil
	}
	if lastPingErr != nil {
		return nil, fmt.Errorf("no payment gateway available in priority list: %w", lastPingErr)
	}

	return nil, fmt.Errorf("no payment gateway available in priority list")
}
