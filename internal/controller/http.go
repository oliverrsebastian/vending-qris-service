package controller

import (
	"encoding/json"
	"net/http"
	"vending-qris-service/internal/logger"

	"vending-qris-service/internal/domain"
	"vending-qris-service/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"moul.io/chizap"
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
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(RequestID)
	r.Use(chizap.New(logger.Get(), &chizap.Opts{
		WithReferer:   true,
		WithUserAgent: true,
	}))

	r.Get("/health", h.healthHandler)
	r.Post("/v1/payments/qris/dynamic", Handle(h.postDynamicQRIS))

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

func (h HTTPServer) postDynamicQRIS(r *http.Request) (Response, error) {
	var req domain.DynamicQRISRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return Response{Code: http.StatusBadRequest}, err
	}

	resp, err := h.qris.GenerateDynamicQRIS(r.Context(), req)
	if err != nil {
		return Response{Code: http.StatusInternalServerError}, err
	}

	return Response{Code: http.StatusOK, Data: resp}, nil
}
