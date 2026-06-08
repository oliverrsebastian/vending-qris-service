package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"vending-qris-client/internal/auth"
	"vending-qris-client/internal/config"
)

type Client struct {
	baseURL    string
	authKey    string
	httpClient *http.Client
}

func New(cfg config.Config) *Client {
	return &Client{
		baseURL: cfg.BaseURL,
		authKey: cfg.AuthKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	DB      bool   `json:"db"`
}

type ProductDetail struct {
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	ItemPrice float64 `json:"item_price"`
}

type GenerateDynamicQRISRequest struct {
	Products      []ProductDetail `json:"products"`
	Description   string          `json:"description,omitempty"`
	InvoiceNumber string          `json:"invoice_number,omitempty"`
}

type APIEnvelope struct {
	Data   json.RawMessage `json:"data"`
	Code   int             `json:"code"`
	Status string          `json:"status"`
	Error  string          `json:"error,omitempty"`
}

type DynamicQRISResponse struct {
	QRString    string `json:"qr_string"`
	ReferenceID string `json:"reference_id"`
	StatusCode  int    `json:"status_code,omitempty"`
	GatewayUsed string `json:"gateway_used,omitempty"`
}

func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	var out HealthResponse
	if err := c.getJSON(ctx, "/health", false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type PaymentGatewaysResponse struct {
	Priority []string `json:"priority"`
	Active   string   `json:"active"`
}

func (c *Client) ListPaymentGateways(ctx context.Context) (*PaymentGatewaysResponse, error) {
	var envelope APIEnvelope
	if err := c.do(ctx, http.MethodGet, "/v1/admin/payment-gateways", true, nil, &envelope); err != nil {
		return nil, err
	}
	if envelope.Code != http.StatusOK {
		return nil, fmt.Errorf("list gateways failed: code=%d error=%s", envelope.Code, envelope.Error)
	}

	var resp PaymentGatewaysResponse
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, &resp); err != nil {
			return nil, err
		}
	}
	return &resp, nil
}

func (c *Client) GenerateDynamicQRIS(ctx context.Context, req GenerateDynamicQRISRequest) (*DynamicQRISResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	var envelope APIEnvelope
	if err := c.do(ctx, http.MethodPost, "/v1/payments/qris/dynamic", false, body, &envelope); err != nil {
		return nil, err
	}
	if envelope.Code != http.StatusOK {
		return nil, fmt.Errorf("generate qris failed: code=%d error=%s", envelope.Code, envelope.Error)
	}

	var resp DynamicQRISResponse
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, &resp); err != nil {
			return nil, err
		}
	}
	return &resp, nil
}

func (c *Client) getJSON(ctx context.Context, path string, requireAuth bool, out any) error {
	return c.do(ctx, http.MethodGet, path, requireAuth, nil, out)
}

func (c *Client) do(ctx context.Context, method, path string, requireAuth bool, body []byte, out any) error {
	if requireAuth && c.authKey == "" {
		return fmt.Errorf("PAYMENT_SERVICE_AUTH_KEY is required")
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	auth.ApplyAuthorization(req, c.authKey)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	payload, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return fmt.Errorf("decode response: %w (body=%s)", err, string(payload))
	}
	return nil
}
