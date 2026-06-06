package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"vending-qris-service/internal/logger"
	"vending-qris-service/internal/request"
	"vending-qris-service/internal/response"
	"vending-qris-service/utilities"

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

func (u *qrisUsecase) GenerateDynamicQRIS(ctx context.Context, req request.DynamicQRISRequest) (*response.DynamicQRISResponse, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	comm, err := u.saver.Save(ctx, &domain.PaymentGatewayCommunication{
		Operation:        opGenerateDynamicQRIS,
		RequestJSON:      reqBytes,
		RequestTimestamp: time.Now(),
	})
	if err != nil {
		return nil, err
	}

	gw, err := u.resolve.Resolve(ctx)
	if err != nil {
		return nil, err
	}

	var resp *response.DynamicQRISResponse

	if err := utilities.Retry(3, 1*time.Second, func() error {
		resp, err = gw.GenerateDynamicQRIS(ctx, req)
		if err != nil {
			logger.Error("error when generate dynamic QRIS with req: %+v, cause: %v", req, err)

			return err
		}

		if resp == nil {
			logger.Error("response is nil with req: %+v", req)

			return err
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

	comm.GatewayName = gw.Name()
	comm.ResponseJSON = respBytes
	comm.ResponseTimestamp = time.Now()
	comm.ResponseStatus = resp.StatusCode

	if _, saveErr := u.saver.Save(ctx, comm); saveErr != nil {
		return nil, saveErr
	}

	resp.GatewayUsed = gw.Name()
	return resp, nil
}
