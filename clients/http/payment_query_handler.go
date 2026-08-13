package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/I-Frostbyte/rvpay-go/clients/payments"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PaymentQueryHandler exposes the GHL Custom Payment Provider query endpoint.
type PaymentQueryHandler struct {
	service *payments.Service
	logger  zerolog.Logger
}

// NewPaymentQueryHandler creates a new payment query HTTP handler.
func NewPaymentQueryHandler(service *payments.Service, logger zerolog.Logger) *PaymentQueryHandler {
	return &PaymentQueryHandler{
		service: service,
		logger:  logger,
	}
}

// Query handles HighLevel payment query requests.
//
// Route: POST /payments/custom-provider/query
// Body: {"type":"verify","transactionId":"...","apiKey":"...","chargeId":"...","subscriptionId":"..."}
func (h *PaymentQueryHandler) Query(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read payment query request body")
		http.Error(w, "could not read request body", http.StatusBadRequest)
		return
	}

	var req payments.QueryRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error().Err(err).Msg("failed to parse payment query request")
		http.Error(w, "invalid payment query request", http.StatusBadRequest)
		return
	}

	resp, err := h.service.HandleQuery(r.Context(), &req)
	if err != nil {
		h.writeQueryError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// writeQueryError maps domain errors to appropriate HTTP status codes without
// exposing internal details, API keys, or transaction data.
func (h *PaymentQueryHandler) writeQueryError(w http.ResponseWriter, err error) {
	code := status.Code(err)

	switch code {
	case codes.InvalidArgument:
		http.Error(w, "invalid payment query request", http.StatusBadRequest)
	case codes.Unauthenticated:
		http.Error(w, "invalid provider API key", http.StatusUnauthorized)
	case codes.NotFound:
		http.Error(w, "transaction not found", http.StatusNotFound)
	case codes.FailedPrecondition:
		http.Error(w, "payment query not permitted", http.StatusBadRequest)
	default:
		h.logger.Error().Err(err).Msg("payment query processing failed")
		http.Error(w, "payment query processing failed", http.StatusInternalServerError)
	}
}
