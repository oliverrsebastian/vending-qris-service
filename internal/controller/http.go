package controller

import (
	"encoding/json"
	"net/http"
	"vending-qris-service/internal/logger"

	"vending-qris-service/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"moul.io/chizap"
)

// Deps bundles HTTP dependencies (keeps constructor readable as the surface grows).
type Deps struct {
	Health   HealthChecker
	QRIS     usecase.QRIS
	Retry    usecase.CommunicationRetry
	Routing  usecase.GatewayRouting
	Callback usecase.PaymentCallback
	AuthKey  string
}

type HTTPServer struct {
	health   HealthChecker
	qris     usecase.QRIS
	retry    usecase.CommunicationRetry
	routing  usecase.GatewayRouting
	callback usecase.PaymentCallback
	mux      chi.Router
	authKey  string
}

func NewHTTPServer(d Deps) HTTPServer {
	h := HTTPServer{
		health:   d.Health,
		qris:     d.QRIS,
		retry:    d.Retry,
		routing:  d.Routing,
		callback: d.Callback,
		authKey:  d.AuthKey,
	}

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(RequestID)
	r.Use(chizap.New(logger.Get(), &chizap.Opts{
		WithReferer:   true,
		WithUserAgent: true,
	}))

	r.Get("/health", h.healthHandler)

	r.Post("/v1/payments/qris/dynamic", Handle(h.postDynamicQRIS))
	r.Post("/v1/payments/check/{transaction_id}", Handle(h.checkPaymentResult))
	r.Post("/v1/payments/cancel/{transaction_id}", Handle(h.cancelPaymentByTransactionID))

	h.registerCallbackRoutes(r)
	h.registerAdminRoutes(r)

	h.mux = r
	return h
}

func (h HTTPServer) Handler() http.Handler { return h.mux }

func (h HTTPServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	dbOK := h.health != nil && h.health.Ping(r.Context()) == nil
	body := map[string]any{
		"status":  "ok",
		"service": "vending-qris-service",
		"db":      dbOK,
	}
	if !dbOK {
		w.WriteHeader(http.StatusServiceUnavailable)
		body["status"] = "degraded"
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}
