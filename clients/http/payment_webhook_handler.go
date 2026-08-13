package http

import (
	"io"
	"net/http"

	"github.com/I-Frostbyte/rvpay-go/clients/payments"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PaymentWebhookHandler exposes the GHL Custom Payment Provider webhook
// endpoint. This is distinct from the Marketplace application webhook
// (/webhooks/highlevel); it handles payment-related events.
type PaymentWebhookHandler struct {
	service *payments.Service
	logger  zerolog.Logger
}

// NewPaymentWebhookHandler creates a new payment webhook HTTP handler.
func NewPaymentWebhookHandler(service *payments.Service, logger zerolog.Logger) *PaymentWebhookHandler {
	return &PaymentWebhookHandler{
		service: service,
		logger:  logger,
	}
}

// Payment handles HighLevel payment-provider webhook deliveries.
//
// Route: POST /payments/custom-provider/webhook
// Body: HighLevel payment event payload (e.g. payment.captured)
func (h *PaymentWebhookHandler) Payment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read payment webhook request body")
		http.Error(w, "could not read request body", http.StatusBadRequest)
		return
	}

	err = h.service.HandleWebhook(r.Context(), body)
	if err != nil {
		h.writeWebhookError(w, err)
		return
	}

	// Acknowledge the event quickly. Processing is synchronous but bounded.
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// writeWebhookError maps domain errors to appropriate HTTP status codes
// without exposing internal details, API keys, or payload contents.
func (h *PaymentWebhookHandler) writeWebhookError(w http.ResponseWriter, err error) {
	code := status.Code(err)

	switch code {
	case codes.InvalidArgument:
		http.Error(w, "invalid payment webhook request", http.StatusBadRequest)
	case codes.NotFound:
		http.Error(w, "payment webhook target not found", http.StatusNotFound)
	case codes.AlreadyExists:
		// Duplicate event: acknowledge as success so the provider stops retrying.
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	case codes.FailedPrecondition:
		http.Error(w, "payment webhook not permitted", http.StatusBadRequest)
	default:
		h.logger.Error().Err(err).Msg("payment webhook processing failed")
		http.Error(w, "payment webhook processing failed", http.StatusInternalServerError)
	}
}
