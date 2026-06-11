package domain

import (
	"context"
	"vending-qris-service/internal/request"
	"vending-qris-service/internal/response"
)

// PaymentGateway is implemented by each provider (Midtrans, Xendit, bank APIs, etc.).
// Swap implementations via config without changing use cases.
type PaymentGateway interface {
	Name() string
	// Ping is a cheap liveness check (TCP/HTTP health to provider, etc.). Return nil if this instance may serve traffic.
	Ping(ctx context.Context) error
	GenerateDynamicQRIS(ctx context.Context, req string) (*response.DynamicQRISResponse, error)
	CheckPaymentStatus(ctx context.Context, req string) (*response.PaymentStatusResponse, error)
	CancelPayment(ctx context.Context, req string) (*response.CancelPaymentResponse, error)
	CreatePayload(ctx context.Context, req request.DynamicQRISRequest) (any, error)
	CancelPayload(ctx context.Context, req request.CancelQRIS) (any, error)
}
