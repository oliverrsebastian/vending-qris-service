package gateway

import (
	"testing"
)

func TestNewCallbackHandlerStub(t *testing.T) {
	h, err := NewCallbackHandler("stub")
	if err != nil {
		t.Fatalf("NewCallbackHandler: %v", err)
	}
	if h.Name() != "stub" {
		t.Fatalf("name=%q", h.Name())
	}
}

func TestNewCallbackHandlerUnknown(t *testing.T) {
	if _, err := NewCallbackHandler("midtrans"); err == nil {
		t.Fatal("expected error for unimplemented gateway")
	}
}
