package response

import "encoding/json"

// DynamicQRISResponse is what a gateway returns after creating a dynamic QRIS (e.g. QR string, deep link).
type DynamicQRISResponse struct {
	QRString       string          `json:"qr_string"`
	RawPayload     json.RawMessage `json:"raw_payload,omitempty"`
	ExpirationTime string          `json:"expiration_time,omitempty"`
	ReferenceID    string          `json:"reference_id"`
	// StatusCode is a gateway-specific outcome (HTTP code, business code, or e.g. "pending" when async).
	StatusCode int `json:"status_code,omitempty"`
	// GatewayUsed is the provider id chosen for this call (from priority + availability).
	GatewayUsed string `json:"gateway_used,omitempty"`
}

type CheckPaymentResponse struct {
	IsPaid bool `json:"is_paid"`
}

type CancelPaymentResponse struct {
	Status string `json:"status"`
}

func (r CancelPaymentResponse) IsSuccess() bool {
	return r.Status == "SUCCESS"
}
