package domain

import (
	"context"
	"encoding/json"
)

// DynamicQRISRequest holds parameters for generating a merchant-presented (dynamic) QRIS payload.
// Field names align with common gateway / QRIS data elements; extend as you add real providers.
type DynamicQRISRequest struct {
	TransactionLedgerID int64  `json:"transaction_ledger_id"`
	AmountMinor         int64  `json:"amount_minor"` // smallest currency unit (e.g. IDR rupiah as integer)
	Currency            string `json:"currency"`     // e.g. IDR
	MerchantName        string `json:"merchant_name"`
	CustomerName        string `json:"customer_name"`
	ReferenceID         string `json:"reference_id"`
	Description         string `json:"description"`
	InvoiceNumber       string `json:"invoice_number,omitempty"`
}

// DynamicQRISResponse is what a gateway returns after creating a dynamic QRIS (e.g. QR string, deep link).
type DynamicQRISResponse struct {
	QRString       string          `json:"qr_string"`
	RawPayload     json.RawMessage `json:"raw_payload,omitempty"`
	ExpirationTime string          `json:"expiration_time,omitempty"`
	ReferenceID    string          `json:"reference_id"`
	// StatusCode is a gateway-specific outcome (HTTP code, business code, or e.g. "pending" when async).
	StatusCode string `json:"status_code,omitempty"`
	// GatewayUsed is the provider id chosen for this call (from priority + availability).
	GatewayUsed string `json:"gateway_used,omitempty"`
}

// PaymentStatusCheckInput is passed when re-querying the provider after an initial communication returned a retryable status.
type PaymentStatusCheckInput struct {
	ReferenceID         string
	TransactionLedgerID int64
	Operation           string
	PollAttempt         int
	RequestJSON         []byte
}

// PaymentStatusResult is the provider response for a status poll (same StatusCode convention as DynamicQRISResponse).
type PaymentStatusResult struct {
	ReferenceID string          `json:"reference_id"`
	StatusCode  string          `json:"status_code"`
	RawPayload  json.RawMessage `json:"raw_payload,omitempty"`
}

// PaymentGateway is implemented by each provider (Midtrans, Xendit, bank APIs, etc.).
// Swap implementations via config without changing use cases.
type PaymentGateway interface {
	Name() string
	// Ping is a cheap liveness check (TCP/HTTP health to provider, etc.). Return nil if this instance may serve traffic.
	Ping(ctx context.Context) error
	GenerateDynamicQRIS(ctx context.Context, req DynamicQRISRequest) (*DynamicQRISResponse, error)
	CheckPaymentStatus(ctx context.Context, in PaymentStatusCheckInput) (*PaymentStatusResult, error)
}
