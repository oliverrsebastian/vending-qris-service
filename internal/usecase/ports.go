package usecase

import (
	"context"
	"net/http"

	"vending-qris-service/internal/request"
	"vending-qris-service/internal/response"
)

// QRIS generates dynamic QRIS payments and records gateway communication.
type QRIS interface {
	GenerateDynamicQRIS(ctx context.Context, req request.DynamicQRISRequest) (*response.DynamicQRISResponse, error)
	CheckPaymentByTransactionID(ctx context.Context, transactionID string) (*response.CheckPaymentResponse, error)
	CancelPaymentByTransactionID(ctx context.Context, transactionID string) (*response.CancelPaymentResponse, error)
}

// CommunicationRetry re-queries the payment gateway for retryable communication rows.
type CommunicationRetry interface {
	PollRetryable(ctx context.Context, force bool)
}

// GatewayRouting manages persisted gateway priority and admin failover.
type GatewayRouting interface {
	ListAndActive(ctx context.Context) (priority []string, active string, err error)
	SetPriority(ctx context.Context, gateways []string) error
	FailoverRotate(ctx context.Context) (priority []string, active string, err error)
}

// PaymentCallback handles inbound payment-gateway webhooks.
type PaymentCallback interface {
	HandleGatewayCallback(ctx context.Context, gatewayName string, headers http.Header, body []byte) error
}
