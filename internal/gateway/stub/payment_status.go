package stub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"vending-qris-service/internal/response"
)

type paymentStatusCheckRequest struct {
	ReferenceID string `json:"reference_id"`
	RequestJSON string `json:"request_json,omitempty"`
}

func (Gateway) CheckPaymentStatus(_ context.Context, req string) (*response.PaymentStatusResponse, error) {
	var payload paymentStatusCheckRequest
	if err := json.Unmarshal([]byte(req), &payload); err != nil {
		return nil, fmt.Errorf("stub status check: invalid json: %w", err)
	}

	referenceID := strings.TrimSpace(payload.ReferenceID)
	if referenceID == "" {
		return nil, fmt.Errorf("stub status check: reference_id is required")
	}

	status := "PENDING"
	statusCode := http.StatusAccepted
	if strings.Contains(strings.ToUpper(referenceID), "PAID") {
		status = "PAID"
		statusCode = http.StatusOK
	}

	qr := fmt.Sprintf("000201010212STUB|status|%s", referenceID)
	return &response.PaymentStatusResponse{
		ReferenceID: referenceID,
		Status:      status,
		StatusCode:  statusCode,
		QRString:    qr,
		RawPayload:  json.RawMessage(req),
	}, nil
}
