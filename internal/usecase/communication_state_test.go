package usecase

import (
	"net/http"
	"testing"

	"vending-qris-service/internal/domain"
)

func TestClassifyCommunication(t *testing.T) {
	tests := []struct {
		name string
		txn  *domain.Transaction
		comm *domain.PaymentGatewayCommunication
		want communicationState
	}{
		{
			name: "empty response",
			txn:  &domain.Transaction{Status: "PENDING"},
			comm: &domain.PaymentGatewayCommunication{ResponseStatus: http.StatusOK},
			want: communicationStateEmptyResponse,
		},
		{
			name: "paid transaction",
			txn:  &domain.Transaction{Status: "PAID"},
			comm: &domain.PaymentGatewayCommunication{
				ResponseJSON:   `{"status_code":200,"qr_string":"qr"}`,
				ResponseStatus: http.StatusOK,
			},
			want: communicationStateSuccess,
		},
		{
			name: "failed",
			txn:  &domain.Transaction{Status: "PENDING"},
			comm: &domain.PaymentGatewayCommunication{
				ResponseJSON:   `{"status_code":400}`,
				ResponseStatus: http.StatusBadRequest,
			},
			want: communicationStateFailed,
		},
		{
			name: "pending payment",
			txn:  &domain.Transaction{Status: "PENDING"},
			comm: &domain.PaymentGatewayCommunication{
				ResponseJSON:   `{"status_code":200,"qr_string":"qr","reference_id":"stub-ref-1"}`,
				ResponseStatus: http.StatusOK,
			},
			want: communicationStatePending,
		},
		{
			name: "paid in stored response",
			txn:  &domain.Transaction{Status: "PENDING"},
			comm: &domain.PaymentGatewayCommunication{
				ResponseJSON:   `{"status":"PAID","status_code":200,"qr_string":"qr"}`,
				ResponseStatus: http.StatusOK,
			},
			want: communicationStateSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifyCommunication(tt.txn, tt.comm); got != tt.want {
				t.Fatalf("got %d want %d", got, tt.want)
			}
		})
	}
}
