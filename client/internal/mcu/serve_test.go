package mcu

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"vending-qris-client/internal/api"
	"vending-qris-client/internal/config"
)

func TestServeValidRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"qr_string":"000201010212","reference_id":"ref-1"},"code":200,"status":"ok"}`))
	}))
	defer srv.Close()

	cfg := config.Config{BaseURL: srv.URL}
	client := api.New(cfg)

	in := strings.NewReader(`{"product":"Water","quantity":1,"price":5000}` + "\n")
	var out bytes.Buffer

	if err := Serve(context.Background(), in, &out, client); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	var resp Response
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok, got error=%s", resp.Error)
	}
	if resp.QRString == "" {
		t.Fatal("expected qr_string")
	}
	if resp.QR == nil || resp.QR.Size == 0 {
		t.Fatal("expected rendered qr image")
	}
}

func TestServeInvalidJSON(t *testing.T) {
	cfg := config.Config{BaseURL: "http://localhost:1"}
	client := api.New(cfg)

	in := strings.NewReader("{not-json\n")
	var out bytes.Buffer

	if err := Serve(context.Background(), in, &out, client); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	var resp Response
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.OK {
		t.Fatal("expected failure for invalid json")
	}
}
