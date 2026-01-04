package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Oidiral/shipment-customer-system/internal/shipment/service"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	svc    *service.Service
	logger *slog.Logger
}

func NewHandler(svc *service.Service, logger *slog.Logger) *Handler {
	return &Handler{
		svc:    svc,
		logger: logger,
	}
}

type CreateShipmentRequest struct {
	Route    string          `json:"route"`
	Price    float64         `json:"price"`
	Customer CustomerRequest `json:"customer"`
}

type CustomerRequest struct {
	IDN string `json:"idn"`
}

type CreateShipmentResponse struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	CustomerID string `json:"customerId"`
}

type GetShipmentResponse struct {
	ID         string  `json:"id"`
	Route      string  `json:"route"`
	Price      float64 `json:"price"`
	Status     string  `json:"status"`
	CustomerID string  `json:"customerId"`
	CreatedAt  string  `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) CreateShipment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	h.logger.Info("CreateShipment called", "trace_id", traceID)

	var req CreateShipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", "error", err, "trace_id", traceID)
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	shipment, err := h.svc.CreateShipment(ctx, service.CreateShipmentRequest{
		req.Route, req.Price, req.Customer.IDN,
	})
	if err != nil {
		h.logger.Error("failed to create shipment", "error", err, "trace_id", traceID)
		h.respondError(w, http.StatusInternalServerError, "failed to create shipment")
		return
	}

	h.respondJSON(w, http.StatusCreated, CreateShipmentResponse{shipment.ID, shipment.Status, shipment.CustomerID})

}

func (h *Handler) GetShipment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/shipments/")
	id := strings.TrimSuffix(path, "/")

	if id == "" {
		h.respondError(w, http.StatusBadRequest, "shipment id is required")
		return
	}

	h.logger.Info("GetShipment request received", "id", id, "trace_id", traceID)

	shipment, err := h.svc.GetShipment(ctx, id)
	if err != nil {
		h.logger.Error("failed to get shipment", "error", err, "id", id, "trace_id", traceID)
		h.respondError(w, http.StatusNotFound, "shipment not found")
		return
	}

	h.respondJSON(w, http.StatusOK, GetShipmentResponse{
		ID:         shipment.ID,
		Route:      shipment.Route,
		Price:      shipment.Price,
		Status:     shipment.Status,
		CustomerID: shipment.CustomerID,
		CreatedAt:  shipment.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, ErrorResponse{Error: message})
}
