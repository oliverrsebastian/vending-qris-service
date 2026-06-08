package worker_test

import (
	"context"
	"testing"
	"time"

	"vending-qris-service/internal/usecase"
	"vending-qris-service/internal/worker"
)

type noopRetry struct {
	calls int
}

func (n *noopRetry) PollRetryable(context.Context, bool) {
	n.calls++
}

func TestRunPaymentCommunicationPoll_respectsContextCancel(t *testing.T) {
	retry := &noopRetry{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		worker.RunPaymentCommunicationPoll(ctx, usecase.RetryPolicy{Enabled: true, IntervalSeconds: 1}, retry)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("worker did not stop after context cancel")
	}
	if retry.calls == 0 {
		t.Fatal("expected at least one poll before cancel")
	}
}
