package usecase

import (
	"context"
	"encoding/json"
	"time"

	"vending-qris-service/internal/domain"
)

type qrisUsecase struct {
	resolve domain.PaymentGatewayResolver
	saver   domain.CommunicationRepository
}

func NewQRISUsecase(resolve domain.PaymentGatewayResolver, saver domain.CommunicationRepository) QRIS {
	return &qrisUsecase{resolve: resolve, saver: saver}
}

const opGenerateDynamicQRIS = "generate_dynamic_qris"

func (u *qrisUsecase) GenerateDynamicQRIS(ctx context.Context, req domain.DynamicQRISRequest) (*domain.DynamicQRISResponse, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	comm, err := u.saver.Save(ctx, &domain.PaymentGatewayCommunication{
		TransactionLedgerID: req.TransactionLedgerID,
		Operation:           opGenerateDynamicQRIS,
		RequestJSON:         reqBytes,
		RequestTimestamp:    time.Now(),
	})
	if err != nil {
		return nil, err
	}

	gw, err := u.resolve.Resolve(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gw.GenerateDynamicQRIS(ctx, req)
	if err != nil && resp == nil {
		return nil, err
	}

	if resp == nil {
		return nil, err
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	status := resp.StatusCode
	if status == "" {
		status = "200"
	}

	comm.GatewayName = gw.Name()
	comm.ResponseJSON = respBytes
	comm.ResponseTimestamp = time.Now()
	comm.ResponseStatus = status

	if _, saveErr := u.saver.Save(ctx, comm); saveErr != nil {
		return nil, saveErr
	}

	resp.GatewayUsed = gw.Name()
	return resp, nil
}
