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

type mockResolver struct {
	gw  domain.PaymentGateway
	err error
}

func (m *mockResolver) Resolve(ctx context.Context) (domain.PaymentGateway, error) {
	_ = ctx
	return m.gw, m.err
}

type mockCommRepo struct {
	created []*domain.PaymentGatewayCommunication
	err     error
}

func (m *mockCommRepo) Create(ctx context.Context, comm *domain.PaymentGatewayCommunication) error {
	_ = ctx
	if m.err != nil {
		return m.err
	}
	m.created = append(m.created, comm)
	return nil
}

func (m *mockCommRepo) ListRetryableByResponseStatus(context.Context, []string, int, int) ([]domain.PaymentGatewayCommunication, error) {
	return nil, nil
}

func (m *mockCommRepo) UpdateAfterStatusPoll(context.Context, int64, []byte, string, time.Time, int) error {
	return nil
}

type stubGateway struct {
	name string
	resp *domain.DynamicQRISResponse
	err  error
}

func (s stubGateway) Name() string               { return s.name }
func (s stubGateway) Ping(context.Context) error { return nil }
func (s stubGateway) GenerateDynamicQRIS(ctx context.Context, req domain.DynamicQRISRequest) (*domain.DynamicQRISResponse, error) {
	_ = ctx
	if s.err != nil {
		return s.resp, s.err
	}
	if s.resp != nil {
		return s.resp, nil
	}
	return &domain.DynamicQRISResponse{
		QRString:    "qr-test",
		ReferenceID: req.ReferenceID,
		StatusCode:  "200",
	}, nil
}
func (s stubGateway) CheckPaymentStatus(_ context.Context, in domain.PaymentStatusCheckInput) (*domain.PaymentStatusResult, error) {
	if in.PollAttempt < 3 {
		return &domain.PaymentStatusResult{
			ReferenceID: in.ReferenceID,
			StatusCode:  "pending",
		}, nil
	}
	return &domain.PaymentStatusResult{
		ReferenceID: in.ReferenceID,
		StatusCode:  "paid",
	}, nil
}

func TestQRISUsecase_GenerateDynamicQRIS_persistsCommunication(t *testing.T) {
	repo := &mockCommRepo{}
	uc := usecase.NewQRISUsecase(
		&mockResolver{gw: stubGateway{name: "stub"}},
		repo,
	)
	req := domain.DynamicQRISRequest{
		TransactionLedgerID: 42,
		AmountMinor:         10000,
		Currency:            "IDR",
		ReferenceID:         "ref-1",
	}
	resp, err := uc.GenerateDynamicQRIS(context.Background(), req)
	if err != nil {
		t.Fatalf("GenerateDynamicQRIS: %v", err)
	}
	if resp.GatewayUsed != "stub" {
		t.Fatalf("gateway_used: got %q want stub", resp.GatewayUsed)
	}
	if len(repo.created) != 1 {
		t.Fatalf("created count: got %d want 1", len(repo.created))
	}
	comm := repo.created[0]
	if comm.GatewayName != "stub" || comm.Operation != "generate_dynamic_qris" {
		t.Fatalf("unexpected comm: %+v", comm)
	}
	if comm.ResponseStatus != "200" {
		t.Fatalf("response_status: got %q", comm.ResponseStatus)
	}
}

func TestQRISUsecase_GenerateDynamicQRIS_resolverError(t *testing.T) {
	uc := usecase.NewQRISUsecase(
		&mockResolver{err: errors.New("no gateway")},
		&mockCommRepo{},
	)
	_, err := uc.GenerateDynamicQRIS(context.Background(), domain.DynamicQRISRequest{
		AmountMinor: 100,
		ReferenceID: "x",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestQRISUsecase_GenerateDynamicQRIS_saveError(t *testing.T) {
	repo := &mockCommRepo{err: errors.New("db down")}
	uc := usecase.NewQRISUsecase(
		&mockResolver{gw: stubGateway{name: "stub"}},
		repo,
	)
	_, err := uc.GenerateDynamicQRIS(context.Background(), domain.DynamicQRISRequest{
		AmountMinor: 100,
		ReferenceID: "x",
	})
	if err == nil || err.Error() != "db down" {
		t.Fatalf("err: %v", err)
	}
}

// Ensure request JSON round-trips in saved communication.
func TestQRISUsecase_requestJSONStored(t *testing.T) {
	repo := &mockCommRepo{}
	uc := usecase.NewQRISUsecase(&mockResolver{gw: stubGateway{name: "stub"}}, repo)
	req := domain.DynamicQRISRequest{AmountMinor: 500, ReferenceID: "ref-json", Currency: "IDR"}
	_, _ = uc.GenerateDynamicQRIS(context.Background(), req)
	var decoded domain.DynamicQRISRequest
	if err := json.Unmarshal(repo.created[0].RequestJSON, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ReferenceID != "ref-json" {
		t.Fatalf("reference_id: %q", decoded.ReferenceID)
	}
}
