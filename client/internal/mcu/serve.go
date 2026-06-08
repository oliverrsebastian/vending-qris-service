package mcu

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"vending-qris-client/internal/api"
	"vending-qris-client/internal/qr"
)

const requestTimeout = 30 * time.Second

// Serve reads JSON requests from in (one per line), calls the payment service,
// renders the returned QR string, and writes JSON responses to out (one per line).
func Serve(ctx context.Context, in io.Reader, out io.Writer, client *api.Client) error {
	scanner := bufio.NewScanner(in)
	// QR payloads can be large; allow bigger lines than the default 64 KiB.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	enc := json.NewEncoder(out)

	for scanner.Scan() {
		line := stringsTrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			if err := enc.Encode(Response{OK: false, Error: fmt.Sprintf("invalid request json: %v", err)}); err != nil {
				return err
			}
			continue
		}

		resp := handleRequest(ctx, client, req)
		if err := enc.Encode(resp); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func handleRequest(parent context.Context, client *api.Client, req Request) Response {
	if req.Product == "" {
		return Response{OK: false, Error: "product is required"}
	}
	if req.Quantity <= 0 {
		return Response{OK: false, Error: "quantity must be greater than 0"}
	}
	if req.Price <= 0 {
		return Response{OK: false, Error: "price must be greater than 0"}
	}

	description := req.Description
	if description == "" {
		description = "vending purchase"
	}

	ctx, cancel := context.WithTimeout(parent, requestTimeout)
	defer cancel()

	apiResp, err := client.GenerateDynamicQRIS(ctx, api.GenerateDynamicQRISRequest{
		Description:   description,
		InvoiceNumber: req.Invoice,
		Products: []api.ProductDetail{
			{
				Name:      req.Product,
				Quantity:  req.Quantity,
				ItemPrice: req.Price,
			},
		},
	})
	if err != nil {
		return Response{OK: false, Error: err.Error()}
	}
	if apiResp.QRString == "" {
		return Response{OK: false, Error: "payment service returned empty qr_string"}
	}

	qrImage, err := qr.Render(apiResp.QRString)
	if err != nil {
		return Response{OK: false, Error: err.Error()}
	}

	return Response{
		OK:          true,
		QRString:    apiResp.QRString,
		ReferenceID: apiResp.ReferenceID,
		GatewayUsed: apiResp.GatewayUsed,
		QR:          qrImage,
	}
}

func stringsTrimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r' || s[start] == '\n') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}
