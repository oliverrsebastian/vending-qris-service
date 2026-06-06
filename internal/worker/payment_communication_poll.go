package worker

import (
	"context"
	"log"
	"time"

	"vending-qris-service/internal/usecase"
)

// RunPaymentCommunicationPoll ticks on interval and asks CommunicationRetry to poll the gateway for retryable rows.
func RunPaymentCommunicationPoll(ctx context.Context, policy usecase.RetryPolicy, uc usecase.CommunicationRetry) {
	interval := time.Duration(policy.IntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 10 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	runPollSafely(uc, ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runPollSafely(uc, ctx)
		}
	}
}

func runPollSafely(uc usecase.CommunicationRetry, ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("payment communication poll panic: %v", r)
		}
	}()
	uc.PollRetryable(ctx, false)
}
