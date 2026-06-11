package stub

import (
	"context"
	"net/http"
	"testing"
)

func TestCheckPaymentStatusPending(t *testing.T) {
	resp, err := Gateway{}.CheckPaymentStatus(context.Background(), `{"reference_id":"stub-ref-1"}`)
	if err != nil {
		t.Fatalf("CheckPaymentStatus: %v", err)
	}
	if resp.Status != "PENDING" {
		t.Fatalf("status=%q", resp.Status)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status_code=%d", resp.StatusCode)
	}
}

func TestCheckPaymentStatusPaid(t *testing.T) {
	resp, err := Gateway{}.CheckPaymentStatus(context.Background(), `{"reference_id":"stub-ref-PAID"}`)
	if err != nil {
		t.Fatalf("CheckPaymentStatus: %v", err)
	}
	if resp.Status != "PAID" {
		t.Fatalf("status=%q", resp.Status)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status_code=%d", resp.StatusCode)
	}
}
