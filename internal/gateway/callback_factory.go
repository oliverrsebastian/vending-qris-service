package gateway

import (
	"fmt"

	"vending-qris-service/internal/domain"
	"vending-qris-service/internal/gateway/stub"
)

// NewCallbackHandler returns the webhook handler for a provider path segment (stub, midtrans, xendit, ...).
func NewCallbackHandler(name string) (domain.PaymentGatewayCallbackHandler, error) {
	switch name {
	case "stub":
		return stub.CallbackHandler{}, nil
	default:
		return nil, fmt.Errorf("unknown callback handler %q", name)
	}
}
