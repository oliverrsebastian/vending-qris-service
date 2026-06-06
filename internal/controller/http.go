package controller

import (
	"encoding/json"
	"net/http"

	"payment-service/internal/domain"
	"payment-service/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Deps bundles HTTP dependencies (keeps constructor readable as the surface grows).
type Deps struct {
	Health  HealthChecker
	QRIS    usecase.QRIS
	Retry   usecase.CommunicationRetry
	Routing usecase.GatewayRouting
}

type HTTPServer struct {
	health  HealthChecker
	qris    usecase.QRIS
	retry   usecase.CommunicationRetry
	routing usecase.GatewayRouting
	mux     chi.Router
}

func NewHTTPServer(d Deps) HTTPServer {
	h := HTTPServer{
		health:  d.Health,
		qris:    d.QRIS,
		retry:   d.Retry,
		routing: d.Routing,
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", h.healthHandler)
	r.Post("/v1/payments/qris/dynamic", h.postDynamicQRIS)

	r.Route("/v1/admin", func(r chi.Router) {
		r.Get("/payment-gateways", h.getAdminPaymentGateways)
		r.Put("/payment-gateways/priority", h.putAdminPaymentGatewayPriority)
		r.Post("/payment-gateways/failover", h.postAdminPaymentGatewayFailover)
		r.Post("/payment-communications/poll", h.postAdminPaymentCommunicationsPoll)
	})

	h.mux = r
	return h
}

func (h HTTPServer) Handler() http.Handler { return h.mux }

func (h HTTPServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	dbOK := h.health != nil && h.health.Ping(r.Context()) == nil
	body := map[string]any{
		"status":  "ok",
		"service": "payment-service",
		"db":      dbOK,
	}
	if !dbOK {
		w.WriteHeader(http.StatusServiceUnavailable)
		body["status"] = "degraded"
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

func (h HTTPServer) postDynamicQRIS(w http.ResponseWriter, r *http.Request) {
	var req domain.DynamicQRISRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Currency == "" {
		req.Currency = "IDR"
	}
	if req.ReferenceID == "" {
		writeJSONError(w, http.StatusBadRequest, "reference_id is required")
		return
	}
	if req.AmountMinor <= 0 {
		writeJSONError(w, http.StatusBadRequest, "amount_minor must be positive")
		return
	}

	resp, err := h.qris.GenerateDynamicQRIS(r.Context(), req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h HTTPServer) getAdminPaymentGateways(w http.ResponseWriter, r *http.Request) {
	priority, active, err := h.routing.ListAndActive(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"priority": priority,
		"active":   active,
	})
}

type putGatewayPriorityBody struct {
	Gateways []string `json:"gateways"`
}

func (h HTTPServer) putAdminPaymentGatewayPriority(w http.ResponseWriter, r *http.Request) {
	var body putGatewayPriorityBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.routing.SetPriority(r.Context(), body.Gateways); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	priority, active, err := h.routing.ListAndActive(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"priority": priority,
		"active":   active,
	})
}

func (h HTTPServer) postAdminPaymentGatewayFailover(w http.ResponseWriter, r *http.Request) {
	priority, active, err := h.routing.FailoverRotate(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"priority": priority,
		"active":   active,
		"note":     "rotated priority: former head moved to tail; active is first Ping-healthy gateway in the new order",
	})
}

func (h HTTPServer) postAdminPaymentCommunicationsPoll(w http.ResponseWriter, r *http.Request) {
	if h.retry == nil {
		writeJSONError(w, http.StatusInternalServerError, "retry use case not configured")
		return
	}
	h.retry.PollRetryable(r.Context(), true)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
