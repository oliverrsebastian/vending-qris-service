package worker_test

import (
	"context"
	"testing"
	"time"

	"vending-qris-service/internal/domain"
	"vending-qris-service/internal/usecase"
	"vending-qris-service/internal/worker"
)

type noopCommRepo struct{}

func (noopCommRepo) Create(context.Context, *domain.PaymentGatewayCommunication) error { return nil }
func (noopCommRepo) ListRetryableByResponseStatus(context.Context, []string, int, int) ([]domain.PaymentGatewayCommunication, error) {
	return nil, nil
}
func (noopCommRepo) UpdateAfterStatusPoll(context.Context, int64, []byte, string, time.Time, int) error {
	return nil
}

type stubGW struct{}

func (stubGW) Name() string               { return "stub" }
func (stubGW) Ping(context.Context) error { return nil }
func (stubGW) GenerateDynamicQRIS(context.Context, domain.DynamicQRISRequest) (*domain.DynamicQRISResponse, error) {
	return nil, nil
}
func (stubGW) CheckPaymentStatus(context.Context, domain.PaymentStatusCheckInput) (*domain.PaymentStatusResult, error) {
	return nil, nil
}

func TestRunPaymentCommunicationPoll_respectsContextCancel(t *testing.T) {
	uc := usecase.NewCommunicationRetryUsecase(
		domain.GatewayFactory(func(string) (domain.PaymentGateway, error) { return stubGW{}, nil }),
		noopCommRepo{},
		usecase.RetryPolicy{Enabled: true, IntervalSeconds: 1},
	)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		worker.RunPaymentCommunicationPoll(ctx, usecase.RetryPolicy{Enabled: true, IntervalSeconds: 1}, uc)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("worker did not stop after context cancel")
	}
}
