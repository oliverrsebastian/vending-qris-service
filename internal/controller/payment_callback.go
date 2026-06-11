package controller

import (
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (h HTTPServer) registerCallbackRoutes(r chi.Router) {
	r.Post("/v1/callbacks/{gateway}", Handle(h.postPaymentGatewayCallback))
}

func (h HTTPServer) postPaymentGatewayCallback(r *http.Request) (Response, error) {
	if h.callback == nil {
		return Response{Code: http.StatusInternalServerError}, errCallbackNotConfigured
	}

	gatewayName := strings.TrimSpace(chi.URLParam(r, "gateway"))
	if gatewayName == "" {
		return Response{Code: http.StatusBadRequest}, errMissingCallbackGateway
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return Response{Code: http.StatusBadRequest}, err
	}

	if err := h.callback.HandleGatewayCallback(r.Context(), gatewayName, r.Header, body); err != nil {
		if isUnknownCallbackGateway(err) {
			return Response{Code: http.StatusNotFound}, err
		}
		if isCallbackClientError(err) {
			return Response{Code: http.StatusBadRequest}, err
		}
		if isTransactionNotFound(err) {
			return Response{Code: http.StatusNotFound}, err
		}
		return Response{Code: http.StatusInternalServerError}, err
	}

	return Response{
		Code:   http.StatusOK,
		Status: "ok",
		Data: map[string]string{
			"gateway": gatewayName,
		},
	}, nil
}
