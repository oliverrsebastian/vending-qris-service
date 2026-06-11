package stub

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"vending-qris-service/internal/request"
	"vending-qris-service/internal/response"
)

// Gateway is a placeholder implementation: swap for a real provider that talks to QRIS / acquirer APIs.
type Gateway struct{}

func (Gateway) Name() string { return "stub" }

func (Gateway) Ping(ctx context.Context) error {
	_ = ctx
	return nil
}

func (Gateway) GenerateDynamicQRIS(ctx context.Context, req string) (*response.DynamicQRISResponse, error) {
	_ = ctx
	// Simulated EMVCo-style payload fragment — replace with real gateway signing / QR data.

	qr := fmt.Sprintf("000201010212STUB|%s", req)
	referenceID := fmt.Sprintf("stub-ref-%d", time.Now().UnixNano())
	return &response.DynamicQRISResponse{
		QRString:    qr,
		ReferenceID: referenceID,
		RawPayload:  []byte(req),
		StatusCode:  http.StatusOK,
	}, nil
}

func (Gateway) CreatePayload(ctx context.Context, req request.DynamicQRISRequest) (any, error) {
	return nil, nil
}
