package domain

import (
	"context"
	"net/http"
)

// CallbackOutcome is the normalized result after a gateway-specific callback is parsed.
type CallbackOutcome struct {
	TransactionID int64
	Status        string
	ReferenceID   string
}

// PaymentGatewayCallbackHandler parses provider-specific webhook payloads.
// Each gateway owns its request format; the use case only sees CallbackOutcome.
type PaymentGatewayCallbackHandler interface {
	Name() string
	HandleCallback(ctx context.Context, headers http.Header, body []byte) (*CallbackOutcome, error)
}
