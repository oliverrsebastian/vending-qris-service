package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"payment-service/internal/controller"
	"payment-service/internal/domain"
	"payment-service/internal/usecase"
)

type healthOK struct{}

func (healthOK) Ping(context.Context) error { return nil }

type healthFail struct{}

func (healthFail) Ping(context.Context) error { return errors.New("db down") }

type mockResolver struct {
	gw domain.PaymentGateway
}

func (m *mockResolver) Resolve(context.Context) (domain.PaymentGateway, error) {
	return m.gw, nil
}

type mockCommRepo struct{}

func (mockCommRepo) Create(context.Context, *domain.PaymentGatewayCommunication) error { return nil }
func (mockCommRepo) ListRetryableByResponseStatus(context.Context, []string, int, int) ([]domain.PaymentGatewayCommunication, error) {
	return nil, nil
}
func (mockCommRepo) UpdateAfterStatusPoll(context.Context, int64, []byte, string, time.Time, int) error {
	return nil
}

type mockPriorityRepo struct {
	list []string
}

func (m *mockPriorityRepo) ListOrdered(context.Context) ([]string, error) { return m.list, nil }
func (m *mockPriorityRepo) ReplaceAll(_ context.Context, gateways []string) error {
	m.list = gateways
	return nil
}
func (m *mockPriorityRepo) RotateOnce(context.Context) error { return nil }

type stubGW struct{}

func (stubGW) Name() string                                                  { return "stub" }
func (stubGW) Ping(context.Context) error                                    { return nil }
func (stubGW) GenerateDynamicQRIS(_ context.Context, req domain.DynamicQRISRequest) (*domain.DynamicQRISResponse, error) {
	return &domain.DynamicQRISResponse{QRString: "qr", ReferenceID: req.ReferenceID, StatusCode: "200"}, nil
}
func (stubGW) CheckPaymentStatus(context.Context, domain.PaymentStatusCheckInput) (*domain.PaymentStatusResult, error) {
	return nil, nil
}

func newTestServer() controller.HTTPServer {
	resolver := &mockResolver{gw: stubGW{}}
	priority := &mockPriorityRepo{list: []string{"stub"}}
	comm := mockCommRepo{}
	return controller.NewHTTPServer(controller.Deps{
		Health:  healthOK{},
		QRIS:    usecase.NewQRISUsecase(resolver, comm),
		Retry:   usecase.NewCommunicationRetryUsecase(domain.GatewayFactory(func(string) (domain.PaymentGateway, error) { return stubGW{}, nil }), comm, usecase.RetryPolicy{Enabled: true}),
		Routing: usecase.NewGatewayRoutingUsecase(priority, resolver),
	})
}

func TestHealth_ok(t *testing.T) {
	srv := controller.NewHTTPServer(controller.Deps{Health: healthOK{}})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHealth_degraded(t *testing.T) {
	srv := controller.NewHTTPServer(controller.Deps{Health: healthFail{}})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status: %d", rec.Code)
	}
}

func TestPostDynamicQRIS_success(t *testing.T) {
	srv := newTestServer()
	body, _ := json.Marshal(map[string]any{
		"reference_id": "ref-ctrl",
		"amount_minor": 5000,
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/payments/qris/dynamic", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rec.Code, rec.Body.String())
	}
	var resp domain.DynamicQRISResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.GatewayUsed != "stub" || resp.QRString != "qr" {
		t.Fatalf("resp: %+v", resp)
	}
}

func TestPostDynamicQRIS_validation(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/v1/payments/qris/dynamic", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rec.Code)
	}
}

func TestGetAdminPaymentGateways(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/payment-gateways", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPostAdminPaymentCommunicationsPoll(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/payment-communications/poll", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
}
