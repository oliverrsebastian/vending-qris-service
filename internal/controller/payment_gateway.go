package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h HTTPServer) registerAdminRoutes(r *chi.Mux) {
	r.Route("/v1/admin", func(r chi.Router) {
		r.Get("/payment-gateways", Handle(h.getAdminPaymentGateways))
		r.Put("/payment-gateways/priority", Handle(h.putAdminPaymentGatewayPriority))
		r.Post("/payment-gateways/failover", Handle(h.postAdminPaymentGatewayFailover))
		r.Post("/payment-communications/poll", Handle(h.postAdminPaymentCommunicationsPoll))
	})
}

func (h HTTPServer) getAdminPaymentGateways(r *http.Request) (Response, error) {
	priority, active, err := h.routing.ListAndActive(r.Context())
	if err != nil {
		return Response{Code: http.StatusInternalServerError}, err
	}

	return Response{
		Data: map[string]any{
			"priority": priority,
			"active":   active,
		},
		Code: http.StatusOK,
	}, nil
}

type putGatewayPriorityBody struct {
	Gateways []string `json:"gateways"`
}

func (h HTTPServer) putAdminPaymentGatewayPriority(r *http.Request) (Response, error) {
	var body putGatewayPriorityBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return Response{Code: http.StatusBadRequest}, err
	}

	if err := h.routing.SetPriority(r.Context(), body.Gateways); err != nil {
		return Response{Code: http.StatusBadRequest}, err
	}

	priority, active, err := h.routing.ListAndActive(r.Context())
	if err != nil {
		return Response{Code: http.StatusInternalServerError}, err
	}

	return Response{
		Data: map[string]any{
			"priority": priority,
			"active":   active,
		},
		Code: http.StatusOK,
	}, nil
}

func (h HTTPServer) postAdminPaymentGatewayFailover(r *http.Request) (Response, error) {
	priority, active, err := h.routing.FailoverRotate(r.Context())
	if err != nil {
		return Response{Code: http.StatusInternalServerError}, err
	}

	return Response{
		Data: map[string]any{
			"priority": priority,
			"active":   active,
			"note":     "rotated priority: former head moved to tail; active is first Ping-healthy gateway in the new order",
		},
		Code: http.StatusOK,
	}, nil
}

func (h HTTPServer) postAdminPaymentCommunicationsPoll(r *http.Request) (Response, error) {
	if h.retry == nil {
		return Response{Code: http.StatusInternalServerError}, errors.New("retry use case not configured")
	}

	h.retry.PollRetryable(r.Context(), true)

	return Response{Code: http.StatusOK}, nil
}
