package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"vending-qris-service/internal/domain"
)

// communicationRetryUsecase re-queries the payment gateway for rows whose response_status is still retryable.
type communicationRetryUsecase struct {
	newGateway domain.GatewayFactory
	repo       domain.CommunicationRepository
	policy     RetryPolicy
}

func NewCommunicationRetryUsecase(
	newGateway domain.GatewayFactory,
	repo domain.CommunicationRepository,
	policy RetryPolicy,
) CommunicationRetry {
	return &communicationRetryUsecase{newGateway: newGateway, repo: repo, policy: policy}
}

// PollRetryable processes a batch of retryable communications. If force is true, runs even when retry is disabled (admin / API triggers).
func (u *communicationRetryUsecase) PollRetryable(ctx context.Context, force bool) {
	if !u.policy.Enabled && !force {
		return
	}
	statuses := u.policy.RetryableResponseStatuses
	if len(statuses) == 0 {
		statuses = []string{"pending"}
	}
	maxAttempts := u.policy.MaxPollAttempts
	if maxAttempts <= 0 {
		maxAttempts = 30
	}
	batch := u.policy.BatchLimit
	if batch <= 0 {
		batch = 100
	}
	rows, err := u.repo.ListRetryableByResponseStatus(ctx, statuses, maxAttempts, batch)
	if err != nil {
		log.Printf("payment communication retry: list: %v", err)
		return
	}
	for i := range rows {
		if err := u.pollOne(ctx, &rows[i]); err != nil {
			log.Printf("payment communication retry: id=%d: %v", rows[i].ID, err)
		}
	}
}

func (u *communicationRetryUsecase) pollOne(ctx context.Context, row *domain.PaymentGatewayCommunication) error {
	var req domain.DynamicQRISRequest
	if err := json.Unmarshal(row.RequestJSON, &req); err != nil {
		return err
	}
	if req.ReferenceID == "" {
		return nil
	}

	gw, err := u.newGateway(row.GatewayName)
	if err != nil {
		return err
	}

	nextAttempt := row.PollAttempts + 1
	now := time.Now()

	in := domain.PaymentStatusCheckInput{
		ReferenceID:         req.ReferenceID,
		TransactionLedgerID: row.TransactionLedgerID,
		Operation:           row.Operation,
		PollAttempt:         nextAttempt,
		RequestJSON:         row.RequestJSON,
	}

	result, err := gw.CheckPaymentStatus(ctx, in)
	status := "poll_error"
	var respBytes []byte
	if err != nil {
		respBytes, _ = wrapPollResponse(row.ResponseJSON, nil, err)
	} else if result != nil {
		status = result.StatusCode
		if status == "" {
			status = "ok"
		}
		respBytes, err = wrapPollResponse(row.ResponseJSON, result, nil)
		if err != nil {
			return err
		}
	} else {
		respBytes, _ = wrapPollResponse(row.ResponseJSON, nil, errors.New("gateway returned nil result"))
	}

	return u.repo.UpdateAfterStatusPoll(ctx, row.ID, respBytes, status, now, nextAttempt)
}

func wrapPollResponse(prior []byte, poll *domain.PaymentStatusResult, pollErr error) ([]byte, error) {
	wrapped := map[string]any{}
	if len(prior) > 0 && json.Valid(prior) {
		wrapped["prior_response"] = json.RawMessage(prior)
	}
	if pollErr != nil {
		wrapped["last_status_poll_error"] = pollErr.Error()
	} else if poll != nil {
		wrapped["last_status_poll"] = poll
	}
	return json.Marshal(wrapped)
}
