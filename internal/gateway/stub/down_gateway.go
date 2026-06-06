package stub

import (
	"context"
	"errors"
	"vending-qris-service/internal/request"
	"vending-qris-service/internal/response"
)

// DownGateway always fails Ping so it is skipped when higher-priority gateways are healthy.
type DownGateway struct{}

func (DownGateway) Name() string { return "stub_down" }

func (DownGateway) Ping(ctx context.Context) error {
	_ = ctx
	return errors.New("stub_down: simulated outage")
}

func (DownGateway) GenerateDynamicQRIS(ctx context.Context, req string) (*response.DynamicQRISResponse, error) {
	_ = ctx
	_ = req
	return nil, errors.New("stub_down: unavailable")
}

func (DownGateway) CreatePayload(ctx context.Context, req request.DynamicQRISRequest) (any, error) {
	return nil, nil
}
