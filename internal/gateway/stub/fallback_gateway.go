package stub

import "context"

// FallbackGateway is the same behavior as Gateway but reports a distinct Name() for routing / audit.
type FallbackGateway struct{ Gateway }

func (FallbackGateway) Name() string { return "stub_fallback" }

func (f FallbackGateway) Ping(ctx context.Context) error {
	return f.Gateway.Ping(ctx)
}
