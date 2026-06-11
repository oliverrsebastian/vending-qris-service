package stub

import (
	"context"
	"testing"
)

func TestCallbackHandlerHandleCallback(t *testing.T) {
	body := []byte(`{"transaction_id":42,"status":"paid","reference_id":"ref-1"}`)

	outcome, err := CallbackHandler{}.HandleCallback(context.Background(), nil, body)
	if err != nil {
		t.Fatalf("HandleCallback: %v", err)
	}
	if outcome.TransactionID != 42 {
		t.Fatalf("transaction_id=%d", outcome.TransactionID)
	}
	if outcome.Status != "PAID" {
		t.Fatalf("status=%q", outcome.Status)
	}
	if outcome.ReferenceID != "ref-1" {
		t.Fatalf("reference_id=%q", outcome.ReferenceID)
	}
}

func TestCallbackHandlerInvalidBody(t *testing.T) {
	_, err := CallbackHandler{}.HandleCallback(context.Background(), nil, []byte("not-json"))
	if err == nil {
		t.Fatal("expected error")
	}
}
