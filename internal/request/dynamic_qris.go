package request

import "github.com/shopspring/decimal"

// DynamicQRISRequest holds parameters for generating a merchant-presented (dynamic) QRIS payload.
// Field names align with common gateway / QRIS data elements; extend as you add real providers.
type DynamicQRISRequest struct {
	Products      []ProductDetail `json:"products"`
	Description   string          `json:"description"`
	InvoiceNumber string          `json:"invoice_number,omitempty"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
}

type ProductDetail struct {
	Name      string          `json:"name"`
	Quantity  int             `json:"quantity"`
	ItemPrice decimal.Decimal `json:"item_price"`
}
