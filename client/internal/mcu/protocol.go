package mcu

import "vending-qris-client/internal/qr"

// Request is a single line of JSON from the microcontroller.
type Request struct {
	Product     string  `json:"product"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Description string  `json:"description,omitempty"`
	Invoice     string  `json:"invoice,omitempty"`
}

// Response is written as one JSON line back to the microcontroller.
type Response struct {
	OK          bool      `json:"ok"`
	Error       string    `json:"error,omitempty"`
	QRString    string    `json:"qr_string,omitempty"`
	ReferenceID string    `json:"reference_id,omitempty"`
	GatewayUsed string    `json:"gateway_used,omitempty"`
	QR          *qr.Image `json:"qr,omitempty"`
}
