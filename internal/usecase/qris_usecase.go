package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"vending-qris-service/internal/logger"
	"vending-qris-service/internal/request"
	"vending-qris-service/internal/response"
	"vending-qris-service/utilities"

	"vending-qris-service/internal/domain"
	gwfactory "vending-qris-service/internal/gateway"

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
const opCancelDynamicQRIS = "cancel_dynamic_qris"

type paymentStatusCheckPayload struct {
	ReferenceID string `json:"reference_id"`
	RequestJSON string `json:"request_json"`
}

func (u *qrisUsecase) GenerateDynamicQRIS(ctx context.Context, req request.DynamicQRISRequest) (*response.DynamicQRISResponse, error) {
	products, totalAmount, err := buildProducts(req)
	if err != nil {
		return nil, err
	}

	req.TotalAmount = totalAmount

	gw, err := u.resolve.Resolve(ctx)
	if err != nil {
		return nil, err
	}

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
	payload := string(payloadBytes)

	if invoice := strings.TrimSpace(req.InvoiceNumber); invoice != "" {
		existingTxn, err := u.txnRepo.FindByInvoiceNumber(ctx, invoice)
		if err != nil {
			return nil, err
		}
		if existingTxn != nil {
			comm, err := u.saver.FindLatestByTransactionAndOperation(ctx, existingTxn.ID, opGenerateDynamicQRIS)
			if err != nil {
				return nil, err
			}
			if comm != nil {
				return u.handleExistingCommunication(ctx, existingTxn, comm, gw, payload)
			}
			return u.generateOnExistingTransaction(ctx, existingTxn, gw, payload)
		}
	}

	return u.createAndGenerateQRIS(ctx, req, products, totalAmount, gw, payload)
}

func buildProducts(req request.DynamicQRISRequest) ([]domain.Product, decimal.Decimal, error) {
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
	if len(products) == 0 {
		return nil, decimal.Zero, errors.New("at least one product is required")
	}
	return products, totalAmount, nil
}

func (u *qrisUsecase) handleExistingCommunication(
	ctx context.Context,
	txn *domain.Transaction,
	comm *domain.PaymentGatewayCommunication,
	defaultGW domain.PaymentGateway,
	payload string,
) (*response.DynamicQRISResponse, error) {
	switch classifyCommunication(txn, comm) {
	case communicationStateSuccess, communicationStateFailed:
		return communicationToQRISResponse(comm, defaultGW.Name())
	case communicationStatePending:
		gw, err := u.gatewayForCommunication(ctx, comm, defaultGW)
		if err != nil {
			return nil, err
		}
		return u.checkPaymentStatusAndUpdate(ctx, txn, comm, gw)
	case communicationStateEmptyResponse:
		gw, err := u.gatewayForCommunication(ctx, comm, defaultGW)
		if err != nil {
			return nil, err
		}
		return u.generateQRISAndUpdate(ctx, txn, comm, gw, payload)
	default:
		return u.createAndGenerateQRIS(ctx, request.DynamicQRISRequest{
			InvoiceNumber: txn.InvoiceNumber,
			Products:      productsFromDomain(txn.Products),
		}, txn.Products, txn.Amount, defaultGW, payload)
	}
}

func (u *qrisUsecase) generateOnExistingTransaction(
	ctx context.Context,
	txn *domain.Transaction,
	gw domain.PaymentGateway,
	payload string,
) (*response.DynamicQRISResponse, error) {
	now := time.Now()
	var comm *domain.PaymentGatewayCommunication
	if err := u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		comm, err = u.saver.Save(txCtx, &domain.PaymentGatewayCommunication{
			TransactionID:    txn.ID,
			GatewayName:      gw.Name(),
			Operation:        opGenerateDynamicQRIS,
			RequestJSON:      payload,
			RequestTimestamp: now,
		})
		return err
	}); err != nil {
		return nil, err
	}
	return u.generateQRISAndUpdate(ctx, txn, comm, gw, payload)
}

func (u *qrisUsecase) createAndGenerateQRIS(
	ctx context.Context,
	req request.DynamicQRISRequest,
	products []domain.Product,
	totalAmount decimal.Decimal,
	gw domain.PaymentGateway,
	payload string,
) (*response.DynamicQRISResponse, error) {
	var (
		txn  *domain.Transaction
		comm *domain.PaymentGatewayCommunication
	)
	now := time.Now()

	if err := u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		txn, err = u.txnRepo.Save(txCtx, &domain.Transaction{
			InvoiceNumber: strings.TrimSpace(req.InvoiceNumber),
			Products:      products,
			Status:        "PENDING",
			Amount:        totalAmount,
		})
		if err != nil {
			logger.Error("error creating transaction for req: %+v, cause: %v", req, err)
			return err
		}

		comm, err = u.saver.Save(txCtx, &domain.PaymentGatewayCommunication{
			TransactionID:    txn.ID,
			GatewayName:      gw.Name(),
			Operation:        opGenerateDynamicQRIS,
			RequestJSON:      payload,
			RequestTimestamp: now,
		})
		return err
	}); err != nil {
		return nil, err
	}

	return u.generateQRISAndUpdate(ctx, txn, comm, gw, payload)
}

func (u *qrisUsecase) generateQRISAndUpdate(
	ctx context.Context,
	txn *domain.Transaction,
	comm *domain.PaymentGatewayCommunication,
	gw domain.PaymentGateway,
	payload string,
) (*response.DynamicQRISResponse, error) {
	var resp *response.DynamicQRISResponse
	if err := utilities.Retry(3, 1*time.Second, func() error {
		var err error
		resp, err = gw.GenerateDynamicQRIS(ctx, payload)
		if err != nil {
			logger.Error("error when generate dynamic QRIS for transaction id=%d, cause: %v", txn.ID, err)
			return err
		}
		if resp == nil {
			return fmt.Errorf("gateway returned nil response")
		}
		if resp.StatusCode != http.StatusOK {
			logger.Error("error when generate dynamic QRIS for transaction id=%d, cause: %+v", txn.ID, resp)
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
		return u.saver.UpdateGatewayResponse(
			txCtx,
			comm.ID,
			gw.Name(),
			string(respBytes),
			resp.StatusCode,
			now,
		)
	}); err != nil {
		return nil, err
	}

	resp.GatewayUsed = gw.Name()
	return resp, nil
}

func (u *qrisUsecase) checkPaymentStatusAndUpdate(
	ctx context.Context,
	txn *domain.Transaction,
	comm *domain.PaymentGatewayCommunication,
	gw domain.PaymentGateway,
) (*response.DynamicQRISResponse, error) {
	checkPayload, err := json.Marshal(paymentStatusCheckPayload{
		ReferenceID: referenceIDFromCommunication(comm),
		RequestJSON: comm.RequestJSON,
	})
	if err != nil {
		return nil, err
	}

	var statusResp *response.PaymentStatusResponse
	if err := utilities.Retry(3, 1*time.Second, func() error {
		var err error
		statusResp, err = gw.CheckPaymentStatus(ctx, string(checkPayload))
		if err != nil {
			logger.Error("error when check payment status for transaction id=%d, cause: %v", txn.ID, err)
			return err
		}
		if statusResp == nil {
			return fmt.Errorf("gateway returned nil payment status")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	resp := paymentStatusToQRISResponse(statusResp, gw.Name())
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if err := u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := u.saver.UpdateGatewayResponse(
			txCtx,
			comm.ID,
			gw.Name(),
			string(respBytes),
			statusResp.StatusCode,
			now,
		); err != nil {
			return err
		}

		if statusResp.StatusCode == http.StatusOK {
			txn.Status = strings.ToUpper(strings.TrimSpace(statusResp.Status))
			if txn.Status == "" {
				txn.Status = "PAID"
			}
			_, err := u.txnRepo.Save(txCtx, txn)
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return resp, nil
}

func (u *qrisUsecase) gatewayForCommunication(
	ctx context.Context,
	comm *domain.PaymentGatewayCommunication,
	defaultGW domain.PaymentGateway,
) (domain.PaymentGateway, error) {
	if comm.GatewayName != "" {
		return gwfactory.New(comm.GatewayName)
	}
	if defaultGW != nil {
		return defaultGW, nil
	}
	return u.resolve.Resolve(ctx)
}

func communicationToQRISResponse(comm *domain.PaymentGatewayCommunication, gatewayName string) (*response.DynamicQRISResponse, error) {
	if strings.TrimSpace(comm.ResponseJSON) == "" {
		return nil, errors.New("communication has no gateway response")
	}

	var resp response.DynamicQRISResponse
	if err := json.Unmarshal([]byte(comm.ResponseJSON), &resp); err != nil {
		return nil, fmt.Errorf("decode stored gateway response: %w", err)
	}
	if resp.GatewayUsed == "" {
		resp.GatewayUsed = firstNonEmpty(comm.GatewayName, gatewayName)
	}
	return &resp, nil
}

func paymentStatusToQRISResponse(status *response.PaymentStatusResponse, gatewayName string) *response.DynamicQRISResponse {
	return &response.DynamicQRISResponse{
		QRString:    status.QRString,
		ReferenceID: status.ReferenceID,
		RawPayload:  status.RawPayload,
		StatusCode:  status.StatusCode,
		GatewayUsed: gatewayName,
	}
}

func referenceIDFromCommunication(comm *domain.PaymentGatewayCommunication) string {
	if comm == nil || strings.TrimSpace(comm.ResponseJSON) == "" {
		return ""
	}
	var resp response.DynamicQRISResponse
	if err := json.Unmarshal([]byte(comm.ResponseJSON), &resp); err != nil {
		return ""
	}
	return resp.ReferenceID
}

func productsFromDomain(products []domain.Product) []request.ProductDetail {
	out := make([]request.ProductDetail, 0, len(products))
	for _, p := range products {
		out = append(out, request.ProductDetail{
			Name:      p.Name,
			Quantity:  p.Quantity,
			ItemPrice: p.ItemPrice,
		})
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func (u *qrisUsecase) CheckPaymentByTransactionID(ctx context.Context, transactionID string) (*response.CheckPaymentResponse, error) {
	transactionIDParsed, err := strconv.ParseInt(transactionID, 10, 64)
	if err != nil {
		return nil, err
	}

	txn, err := u.txnRepo.FindByID(ctx, transactionIDParsed)
	if err != nil {
		return nil, err
	}

	if txn == nil {
		return nil, errors.New("transaction not found")
	}

	if txn.Status == "PAID" {
		return &response.CheckPaymentResponse{IsPaid: true}, nil
	}

	gwComm, err := u.saver.FindLatestByTransactionAndOperation(ctx, transactionIDParsed, opGenerateDynamicQRIS)
	if err != nil {
		return nil, err
	}

	if gwComm == nil {
		return nil, errors.New("gateway communication not found")
	}

	if gwComm.ResponseStatus != http.StatusOK {
		return nil, errors.New("gateway communication failed and terminated")
	}

	var resp response.DynamicQRISResponse
	if err := json.Unmarshal([]byte(gwComm.ResponseJSON), &resp); err != nil {
		return nil, err
	}

	// if already successful
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("request rejected by gateway")
	}

	gw, err := u.resolve.ResolveByName(ctx, gwComm.GatewayName)
	if err != nil {
		return nil, err
	}

	checkResp, err := u.checkPaymentStatusAndUpdate(ctx, txn, gwComm, gw)
	if err != nil {
		return nil, err
	}

	if checkResp.StatusCode != http.StatusOK {
		return nil, errors.New("check request rejected by gateway")
	}

	// update to paid status if check response says it's already settled.
	if !txn.IsPaid() {
		txn.UpdateToPaid()
		if _, err := u.txnRepo.Save(ctx, txn); err != nil {
			return nil, err
		}
	}

	return &response.CheckPaymentResponse{IsPaid: true}, nil
}

func (u *qrisUsecase) CancelPaymentByTransactionID(ctx context.Context, transactionID string) (*response.CancelPaymentResponse, error) {
	transactionIDParsed, err := strconv.ParseInt(transactionID, 10, 64)
	if err != nil {
		return nil, err
	}

	txn, err := u.txnRepo.FindByID(ctx, transactionIDParsed)
	if err != nil {
		return nil, err
	}

	if txn == nil {
		return nil, errors.New("transaction not found")
	}

	if txn.IsPaid() {
		return nil, errors.New("transaction is already paid")
	}

	gw, err := u.resolve.Resolve(ctx)
	if err != nil {
		return nil, err
	}

	req := request.CancelQRIS{ReferenceID: transactionID}

	specificPayload, err := gw.CancelPayload(ctx, req)
	if err != nil {
		logger.Error("error creating gateway-specific payload for req: %+v, cause: %v", req, err)
		return nil, err
	}

	payloadBytes, err := json.Marshal(specificPayload)
	if err != nil {
		logger.Error("error creating payload for req: %+v, cause: %v", req, err)
		return nil, err
	}

	payload := string(payloadBytes)

	var comm *domain.PaymentGatewayCommunication

	now := time.Now()

	if err := u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		comm, err = u.saver.Save(txCtx, &domain.PaymentGatewayCommunication{
			TransactionID:    txn.ID,
			GatewayName:      gw.Name(),
			Operation:        opCancelDynamicQRIS,
			RequestJSON:      payload,
			RequestTimestamp: now,
		})
		return err
	}); err != nil {
		return nil, err
	}

	var resp *response.CancelPaymentResponse
	if err := utilities.Retry(3, 1*time.Second, func() error {
		var err error
		resp, err = gw.CancelPayment(ctx, payload)
		if err != nil {
			logger.Error("error when generate dynamic QRIS for transaction id=%d, cause: %v", txn.ID, err)
			return err
		}
		if resp == nil {
			return fmt.Errorf("gateway returned nil response")
		}
		if !resp.IsSuccess() {
			logger.Error("error when generate dynamic QRIS for transaction id=%d, cause: %+v", txn.ID, resp)
			return fmt.Errorf("unexpected status %v", resp.Status)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	if err := u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		return u.saver.UpdateGatewayResponse(
			txCtx,
			comm.ID,
			gw.Name(),
			string(respBytes),
			200,
			now,
		)
	}); err != nil {
		return nil, err
	}

	return resp, nil
}
