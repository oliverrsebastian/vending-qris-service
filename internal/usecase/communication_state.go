package usecase

import (
	"encoding/json"
	"net/http"
	"strings"

	"vending-qris-service/internal/domain"
	"vending-qris-service/internal/response"
)

type communicationState int

const (
	communicationStateNew communicationState = iota
	communicationStateSuccess
	communicationStateFailed
	communicationStatePending
	communicationStateEmptyResponse
)

func classifyCommunication(txn *domain.Transaction, comm *domain.PaymentGatewayCommunication) communicationState {
	if comm == nil {
		return communicationStateNew
	}
	if strings.TrimSpace(comm.ResponseJSON) == "" {
		return communicationStateEmptyResponse
	}
	if comm.ResponseStatus >= http.StatusBadRequest {
		return communicationStateFailed
	}
	if isTerminalTransactionStatus(txn.Status) || storedResponseIsPaid(comm.ResponseJSON) {
		return communicationStateSuccess
	}
	return communicationStatePending
}

func isTerminalTransactionStatus(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "PAID", "SUCCESS", "COMPLETED":
		return true
	default:
		return false
	}
}

func storedResponseIsPaid(responseJSON string) bool {
	var statusResp response.PaymentStatusResponse
	if err := json.Unmarshal([]byte(responseJSON), &statusResp); err != nil {
		return false
	}
	switch strings.ToUpper(strings.TrimSpace(statusResp.Status)) {
	case "PAID", "SUCCESS", "COMPLETED":
		return true
	default:
		return false
	}
}
