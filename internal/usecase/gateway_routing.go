package usecase

import (
	"context"

	"vending-qris-service/internal/domain"
)

type gatewayRoutingUsecase struct {
	priority domain.GatewayPriorityRepository
	resolver domain.PaymentGatewayResolver
}

func NewGatewayRoutingUsecase(
	priority domain.GatewayPriorityRepository,
	resolver domain.PaymentGatewayResolver,
) GatewayRouting {
	return &gatewayRoutingUsecase{priority: priority, resolver: resolver}
}

func (u *gatewayRoutingUsecase) ListAndActive(ctx context.Context) (priority []string, active string, err error) {
	priority, err = u.priority.ListOrdered(ctx)
	if err != nil {
		return nil, "", err
	}

	gw, err := u.resolver.Resolve(ctx)
	if err != nil {
		return priority, "", err
	}

	return priority, gw.Name(), nil
}

func (u *gatewayRoutingUsecase) SetPriority(ctx context.Context, gateways []string) error {
	return u.priority.ReplaceAll(ctx, gateways)
}

// FailoverRotate moves the current head of the priority list to the tail, then returns the list and first Ping-healthy gateway.
func (u *gatewayRoutingUsecase) FailoverRotate(ctx context.Context) (priority []string, active string, err error) {
	if err := u.priority.RotateOnce(ctx); err != nil {
		return nil, "", err
	}
	return u.ListAndActive(ctx)
}
