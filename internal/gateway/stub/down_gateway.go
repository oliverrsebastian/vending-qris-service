package stub

import (
	"context"
	"errors"

	"payment-service/internal/domain"
)

// DownGateway always fails Ping so it is skipped when higher-priority gateways are healthy.
type DownGateway struct{}

func (DownGateway) Name() string { return "stub_down" }

func (DownGateway) Ping(ctx context.Context) error {
	_ = ctx
	return errors.New("stub_down: simulated outage")
}

func (DownGateway) GenerateDynamicQRIS(ctx context.Context, req domain.DynamicQRISRequest) (*domain.DynamicQRISResponse, error) {
	_ = ctx
	_ = req
	return nil, errors.New("stub_down: unavailable")
}

func (DownGateway) CheckPaymentStatus(ctx context.Context, in domain.PaymentStatusCheckInput) (*domain.PaymentStatusResult, error) {
	_ = ctx
	_ = in
	return nil, errors.New("stub_down: unavailable")
}
