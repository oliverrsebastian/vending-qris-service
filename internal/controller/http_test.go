package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"vending-qris-service/internal/controller"
	"vending-qris-service/internal/request"
	"vending-qris-service/internal/response"
	"vending-qris-service/internal/usecase"
)

type healthOK struct{}

func (healthOK) Ping(context.Context) error { return nil }

type healthFail struct{}

func (healthFail) Ping(context.Context) error { return errors.New("db down") }

type mockQRIS struct{}

func (mockQRIS) GenerateDynamicQRIS(context.Context, request.DynamicQRISRequest) (*response.DynamicQRISResponse, error) {
	return &response.DynamicQRISResponse{
		QRString:    "qr-test",
		StatusCode:  http.StatusOK,
		GatewayUsed: "stub",
	}, nil
}

type mockRouting struct{}

func (mockRouting) ListAndActive(context.Context) ([]string, string, error) {
	return []string{"stub"}, "stub", nil
}
func (mockRouting) SetPriority(context.Context, []string) error { return nil }
func (mockRouting) FailoverRotate(context.Context) ([]string, string, error) {
	return []string{"stub"}, "stub", nil
}

func newTestServer(authKey string) controller.HTTPServer {
	return controller.NewHTTPServer(controller.Deps{
		Health:  healthOK{},
		QRIS:    mockQRIS{},
		Routing: mockRouting{},
		AuthKey: authKey,
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
	srv := newTestServer("")
	body, _ := json.Marshal(map[string]any{
		"description": "test",
		"products": []map[string]any{
			{"name": "Water", "quantity": 1, "item_price": 5000},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/payments/qris/dynamic", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetAdminPaymentGateways_requiresAuth(t *testing.T) {
	srv := newTestServer("secret-key")
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/payment-gateways", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status: %d want 401", rec.Code)
	}
}

func TestGetAdminPaymentGateways_withAuth(t *testing.T) {
	srv := newTestServer("secret-key")
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/payment-gateways", nil)
	req.Header.Set("Authorization", "secret-key")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rec.Code, rec.Body.String())
	}
}

// Ensure usecase interfaces remain decoupled from controller wiring.
var (
	_ usecase.QRIS           = mockQRIS{}
	_ usecase.GatewayRouting = mockRouting{}
)
