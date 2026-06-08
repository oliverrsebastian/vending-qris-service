package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"vending-qris-service/internal/logger"
	"vending-qris-service/internal/request"
	"vending-qris-service/internal/response"
	"vending-qris-service/utilities"

	"vending-qris-service/internal/domain"

	"github.com/shopspring/decimal"
)

type qrisUsecase struct {
	resolve    domain.PaymentGatewayResolver
	saver      domain.CommunicationRepository
	txnRepo    domain.TransactionRepository
	transactor domain.Transactor
}

func NewQRISUsecase(
	resolve domain.PaymentGatewayResolver,
	saver domain.CommunicationRepository,
	txnRepository domain.TransactionRepository,
	transactor domain.Transactor,
) QRIS {
	return &qrisUsecase{
		resolve:    resolve,
		saver:      saver,
		txnRepo:    txnRepository,
		transactor: transactor,
	}
}

const opGenerateDynamicQRIS = "generate_dynamic_qris"

func (u *qrisUsecase) GenerateDynamicQRIS(ctx context.Context, req request.DynamicQRISRequest) (*response.DynamicQRISResponse, error) {
	products := make([]domain.Product, 0, len(req.Products))
	totalAmount := decimal.Zero
	for _, product := range req.Products {
		products = append(products, domain.Product{
			Name:      product.Name,
			Quantity:  product.Quantity,
			ItemPrice: product.ItemPrice,
		})
		totalAmount = totalAmount.Add(product.ItemPrice.Mul(decimal.NewFromInt(int64(product.Quantity))))
	}

	gw, err := u.resolve.Resolve(ctx)
	if err != nil {
		return nil, err
	}

	req.TotalAmount = totalAmount

	specificPayload, err := gw.CreatePayload(ctx, req)
	if err != nil {
		logger.Error("error creating gateway-specific payload for req: %+v, cause: %v", req, err)
		return nil, err
	}

	payloadBytes, err := json.Marshal(specificPayload)
	if err != nil {
		logger.Error("error creating payload for req: %+v, cause: %v", req, err)

		return nil, err
	}

	var (
		txn  *domain.Transaction
		comm *domain.PaymentGatewayCommunication
	)
	if err := u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		txn, err = u.txnRepo.Save(txCtx, &domain.Transaction{
			Products: products,
			Status:   "PENDING",
			Amount:   totalAmount,
		})
		if err != nil {
			logger.Error("error creating transaction for req: %+v, cause: %v", req, err)
			return err
		}

		comm, err = u.saver.Save(ctx, &domain.PaymentGatewayCommunication{
			TransactionID:    txn.ID,
			Operation:        opGenerateDynamicQRIS,
			RequestJSON:      string(payloadBytes),
			RequestTimestamp: time.Now(),
		})
		if err != nil {
			return err
		}

		return nil

	}); err != nil {
		return nil, err
	}

	if comm == nil {
		return nil, errors.New("communication is nil")
	}

	var resp *response.DynamicQRISResponse
	if err := utilities.Retry(3, 1*time.Second, func() error {
		resp, err = gw.GenerateDynamicQRIS(ctx, string(payloadBytes))
		if err != nil {
			logger.Error("error when generate dynamic QRIS with req: %+v, cause: %v", req, err)
			return err
		}
		if resp == nil {
			logger.Error("response is nil with req: %+v", req)
			return fmt.Errorf("gateway returned nil response")
		}
		if resp.StatusCode != http.StatusOK {
			logger.Error("error when generate dynamic QRIS with req: %+v, cause: %+v", req, resp)
			return fmt.Errorf("unexpected status code %d", resp.StatusCode)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if err := u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		_, err := u.saver.Save(txCtx, &domain.PaymentGatewayCommunication{
			TransactionID:     txn.ID,
			GatewayName:       gw.Name(),
			Operation:         opGenerateDynamicQRIS,
			RequestJSON:       string(payloadBytes),
			RequestTimestamp:  now,
			ResponseJSON:      string(respBytes),
			ResponseStatus:    resp.StatusCode,
			ResponseTimestamp: now,
		})
		if err != nil {
			logger.Error("error saving communication for transaction id=%d, cause: %v", txn.ID, err)
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	resp.GatewayUsed = gw.Name()
	return resp, nil
}
