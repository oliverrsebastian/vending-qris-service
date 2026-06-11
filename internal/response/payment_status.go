package response

import "encoding/json"

// PaymentStatusResponse is returned when querying a gateway for an existing payment.
type PaymentStatusResponse struct {
	ReferenceID string          `json:"reference_id"`
	Status      string          `json:"status"`
	StatusCode  int             `json:"status_code"`
	QRString    string          `json:"qr_string,omitempty"`
	RawPayload  json.RawMessage `json:"raw_payload,omitempty"`
}
