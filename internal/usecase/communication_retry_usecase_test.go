package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"vending-qris-service/internal/domain"
	"vending-qris-service/internal/usecase"
)

type retryCommRepo struct {
	rows    []domain.PaymentGatewayCommunication
	updated []domain.PaymentGatewayCommunication
	listErr error
	updErr  error
}

func (r *retryCommRepo) Create(context.Context, *domain.PaymentGatewayCommunication) error {
	return nil
}

func (r *retryCommRepo) ListRetryableByResponseStatus(
	_ context.Context,
	_ []string,
	_ int,
	_ int,
) ([]domain.PaymentGatewayCommunication, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.rows, nil
}

func (r *retryCommRepo) UpdateAfterStatusPoll(
	_ context.Context,
	id int64,
	responseJSON []byte,
	responseStatus string,
	_ time.Time,
	pollAttempts int,
) error {
	if r.updErr != nil {
		return r.updErr
	}
	r.updated = append(r.updated, domain.PaymentGatewayCommunication{
		ID:             id,
		ResponseJSON:   responseJSON,
		ResponseStatus: responseStatus,
		PollAttempts:   pollAttempts,
	})
	return nil
}

func TestCommunicationRetryUsecase_PollRetryable_skipsWhenDisabled(t *testing.T) {
	repo := &retryCommRepo{rows: []domain.PaymentGatewayCommunication{{ID: 1}}}
	uc := usecase.NewCommunicationRetryUsecase(
		domain.GatewayFactory(func(string) (domain.PaymentGateway, error) { return stubGateway{name: "stub"}, nil }),
		repo,
		usecase.RetryPolicy{Enabled: false},
	)
	uc.PollRetryable(context.Background(), false)
	if len(repo.updated) != 0 {
		t.Fatal("expected no updates when disabled")
	}
}

func TestCommunicationRetryUsecase_PollRetryable_forceWhenDisabled(t *testing.T) {
	req, _ := json.Marshal(domain.DynamicQRISRequest{ReferenceID: "ref-poll", AmountMinor: 100})
	repo := &retryCommRepo{rows: []domain.PaymentGatewayCommunication{{
		ID:           7,
		GatewayName:  "stub",
		RequestJSON:  req,
		ResponseJSON: []byte(`{"status_code":"pending"}`),
		PollAttempts: 0,
	}}}
	uc := usecase.NewCommunicationRetryUsecase(
		domain.GatewayFactory(func(string) (domain.PaymentGateway, error) { return stubGateway{name: "stub"}, nil }),
		repo,
		usecase.RetryPolicy{Enabled: false, MaxPollAttempts: 30, BatchLimit: 10},
	)
	uc.PollRetryable(context.Background(), true)
	if len(repo.updated) != 1 {
		t.Fatalf("updated: got %d want 1", len(repo.updated))
	}
	if repo.updated[0].PollAttempts != 1 {
		t.Fatalf("poll_attempts: got %d", repo.updated[0].PollAttempts)
	}
}

func TestCommunicationRetryUsecase_PollRetryable_updatesStatus(t *testing.T) {
	req, _ := json.Marshal(domain.DynamicQRISRequest{ReferenceID: "ref-paid", AmountMinor: 100})
	repo := &retryCommRepo{rows: []domain.PaymentGatewayCommunication{{
		ID:           1,
		GatewayName:  "stub",
		RequestJSON:  req,
		PollAttempts: 2, // third poll returns paid from stub
	}}}
	uc := usecase.NewCommunicationRetryUsecase(
		domain.GatewayFactory(func(string) (domain.PaymentGateway, error) { return stubGateway{name: "stub"}, nil }),
		repo,
		usecase.RetryPolicy{
			Enabled:                   true,
			RetryableResponseStatuses: []string{"pending"},
			MaxPollAttempts:           30,
			BatchLimit:                10,
		},
	)
	uc.PollRetryable(context.Background(), false)
	if len(repo.updated) != 1 {
		t.Fatalf("updated: %d", len(repo.updated))
	}
	if repo.updated[0].ResponseStatus != "paid" {
		t.Fatalf("status: got %q want paid", repo.updated[0].ResponseStatus)
	}
}

func TestCommunicationRetryUsecase_PollRetryable_unknownGateway(t *testing.T) {
	req, _ := json.Marshal(domain.DynamicQRISRequest{ReferenceID: "ref-x"})
	repo := &retryCommRepo{rows: []domain.PaymentGatewayCommunication{{
		ID:          2,
		GatewayName: "unknown-gw",
		RequestJSON: req,
	}}}
	uc := usecase.NewCommunicationRetryUsecase(
		domain.GatewayFactory(func(name string) (domain.PaymentGateway, error) {
			return nil, errors.New("unknown")
		}),
		repo,
		usecase.RetryPolicy{Enabled: true, RetryableResponseStatuses: []string{"pending"}, MaxPollAttempts: 5, BatchLimit: 5},
	)
	uc.PollRetryable(context.Background(), false)
	if len(repo.updated) != 0 {
		t.Fatal("expected no update on gateway factory error")
	}
}
