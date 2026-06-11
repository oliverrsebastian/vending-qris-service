package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"vending-qris-service/internal/request"
)

func (h HTTPServer) postDynamicQRIS(r *http.Request) (Response, error) {
	var req request.DynamicQRISRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return Response{Code: http.StatusBadRequest}, err
	}

	resp, err := h.qris.GenerateDynamicQRIS(r.Context(), req)
	if err != nil {
		return Response{Code: http.StatusInternalServerError}, err
	}

	return Response{Code: http.StatusOK, Data: resp}, nil
}

func (h HTTPServer) checkPaymentResult(r *http.Request) (Response, error) {
	transactionID := r.PathValue("transaction_id")
	if transactionID == "" {
		return Response{Code: http.StatusBadRequest}, errors.New("transaction_id is required")
	}

	resp, err := h.qris.CheckPaymentByTransactionID(r.Context(), transactionID)
	if err != nil {
		return Response{Code: http.StatusInternalServerError}, err
	}

	return Response{Code: http.StatusOK, Data: resp}, nil
}

func (h HTTPServer) cancelPaymentByTransactionID(r *http.Request) (Response, error) {
	transactionID := r.PathValue("transaction_id")
	if transactionID == "" {
		return Response{Code: http.StatusBadRequest}, errors.New("transaction_id is required")
	}

	resp, err := h.qris.CancelPaymentByTransactionID(r.Context(), transactionID)
	if err != nil {
		return Response{Code: http.StatusInternalServerError}, err
	}

	return Response{Code: http.StatusOK, Data: resp}, nil
}
