package stub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"vending-qris-service/internal/domain"
)

// CallbackHandler handles POST /v1/callbacks/stub webhooks.
type CallbackHandler struct{}

func (CallbackHandler) Name() string { return "stub" }

type stubCallbackBody struct {
	TransactionID int64  `json:"transaction_id"`
	Status        string `json:"status"`
	ReferenceID   string `json:"reference_id,omitempty"`
}

func (CallbackHandler) HandleCallback(_ context.Context, _ http.Header, body []byte) (*domain.CallbackOutcome, error) {
	var payload stubCallbackBody
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("stub callback: invalid json: %w", err)
	}

	if payload.TransactionID <= 0 {
		return nil, fmt.Errorf("stub callback: transaction_id is required")
	}

	status := strings.TrimSpace(strings.ToUpper(payload.Status))
	if status == "" {
		return nil, fmt.Errorf("stub callback: status is required")
	}

	return &domain.CallbackOutcome{
		TransactionID: payload.TransactionID,
		Status:        status,
		ReferenceID:   strings.TrimSpace(payload.ReferenceID),
	}, nil
}
