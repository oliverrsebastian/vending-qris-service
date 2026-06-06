package controller

import "context"

// HealthChecker reports whether backing infrastructure (e.g. database) is reachable.
type HealthChecker interface {
	Ping(ctx context.Context) error
}
