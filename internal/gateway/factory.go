package gateway

import (
	"fmt"

	"payment-service/internal/domain"
	"payment-service/internal/gateway/stub"
)

// New returns the payment gateway implementation selected by name (env PAYMENT_GATEWAY).
func New(name string) (domain.PaymentGateway, error) {
	switch name {
	case "stub":
		return stub.Gateway{}, nil
	case "stub_fallback":
		return stub.FallbackGateway{}, nil
	case "stub_down":
		return stub.DownGateway{}, nil
	default:
		return nil, fmt.Errorf("unknown payment gateway %q", name)
	}
}
