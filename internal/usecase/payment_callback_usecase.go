package usecase

import (
	"context"
	"net/http"
	"time"

	"vending-qris-service/internal/domain"
	gwfactory "vending-qris-service/internal/gateway"
	"vending-qris-service/internal/logger"
)

const opPaymentCallback = "payment_callback"

type paymentCallbackUsecase struct {
	txnRepo    domain.TransactionRepository
	commRepo   domain.CommunicationRepository
	transactor domain.Transactor
}

func NewPaymentCallbackUsecase(
	txnRepo domain.TransactionRepository,
	commRepo domain.CommunicationRepository,
	transactor domain.Transactor,
) PaymentCallback {
	return &paymentCallbackUsecase{
		txnRepo:    txnRepo,
		commRepo:   commRepo,
		transactor: transactor,
	}
}

func (u *paymentCallbackUsecase) HandleGatewayCallback(
	ctx context.Context,
	gatewayName string,
	headers http.Header,
	body []byte,
) error {
	handler, err := gwfactory.NewCallbackHandler(gatewayName)
	if err != nil {
		return err
	}

	outcome, err := handler.HandleCallback(ctx, headers, body)
	if err != nil {
		return err
	}

	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		txn, err := u.txnRepo.FindByID(txCtx, outcome.TransactionID)
		if err != nil {
			return err
		}

		txn.Status = outcome.Status
		if _, err := u.txnRepo.Save(txCtx, txn); err != nil {
			logger.Error("error updating transaction id=%d after callback, cause: %v", txn.ID, err)
			return err
		}

		now := time.Now()
		_, err = u.commRepo.Save(txCtx, &domain.PaymentGatewayCommunication{
			TransactionID:     txn.ID,
			GatewayName:       handler.Name(),
			Operation:         opPaymentCallback,
			RequestJSON:       string(body),
			RequestTimestamp:  now,
			ResponseStatus:    http.StatusOK,
			ResponseTimestamp: now,
		})
		if err != nil {
			logger.Error("error saving callback communication for transaction id=%d, cause: %v", txn.ID, err)
			return err
		}

		return nil
	})
}
