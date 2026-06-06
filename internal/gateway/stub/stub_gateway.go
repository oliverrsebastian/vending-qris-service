package stub

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"vending-qris-service/internal/domain"
)

// Gateway is a placeholder implementation: swap for a real provider that talks to QRIS / acquirer APIs.
type Gateway struct{}

func (Gateway) Name() string { return "stub" }

func (Gateway) Ping(ctx context.Context) error {
	_ = ctx
	return nil
}

func (Gateway) GenerateDynamicQRIS(ctx context.Context, req domain.DynamicQRISRequest) (*domain.DynamicQRISResponse, error) {
	_ = ctx
	// Simulated EMVCo-style payload fragment — replace with real gateway signing / QR data.
	payload := map[string]any{
		"format":         "qris.dynamic",
		"amount_minor":   req.AmountMinor,
		"currency":       req.Currency,
		"merchant_name":  req.MerchantName,
		"customer_name":  req.CustomerName,
		"reference_id":   req.ReferenceID,
		"description":    req.Description,
		"invoice_number": req.InvoiceNumber,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	qr := fmt.Sprintf("000201010212STUB|%s", string(raw))
	status := "200"
	if strings.Contains(req.ReferenceID, "pending-e2e") {
		status = "pending"
	}
	return &domain.DynamicQRISResponse{
		QRString:    qr,
		RawPayload:  raw,
		ReferenceID: req.ReferenceID,
		StatusCode:  status,
	}, nil
}

func (Gateway) CheckPaymentStatus(ctx context.Context, in domain.PaymentStatusCheckInput) (*domain.PaymentStatusResult, error) {
	_ = ctx
	// Simulates async settlement: still "pending" until the third poll, then "paid".
	if in.PollAttempt < 3 {
		raw, _ := json.Marshal(map[string]any{"stub": "awaiting_acquirer", "attempt": in.PollAttempt})
		return &domain.PaymentStatusResult{
			ReferenceID: in.ReferenceID,
			StatusCode:  "pending",
			RawPayload:  raw,
		}, nil
	}
	raw, _ := json.Marshal(map[string]any{"stub": "settled"})
	return &domain.PaymentStatusResult{
		ReferenceID: in.ReferenceID,
		StatusCode:  "paid",
		RawPayload:  raw,
	}, nil
}
