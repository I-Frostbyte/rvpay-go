package http

import (
	"io"
	"net/http"

	"github.com/I-Frostbyte/rvpay-go/clients/webhooks"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WebhookHandler exposes the provider webhook delivery endpoint.
type WebhookHandler struct {
	service *webhooks.Service
	logger  zerolog.Logger
}

// NewWebhookHandler creates a new webhook HTTP handler.
func NewWebhookHandler(service *webhooks.Service, logger zerolog.Logger) *WebhookHandler {
	return &WebhookHandler{
		service: service,
		logger:  logger,
	}
}

// HighLevel handles HighLevel webhook deliveries.
//
// Route: POST /webhooks/highlevel
// Header: X-GHL-Signature (base64-encoded Ed25519 signature over the raw body)
func (h *WebhookHandler) HighLevel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Read the raw body bytes. The signature is computed over the exact raw
	// body, so we must preserve the original bytes and never re-marshal or
	// transform the JSON before verification.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read webhook request body")
		http.Error(w, "could not read request body", http.StatusBadRequest)
		return
	}

	headers := map[string]string{
		"X-GHL-Signature": r.Header.Get("X-GHL-Signature"),
	}

	err = h.service.ProcessWebhook(r.Context(), "highlevel", headers, body)
	if err != nil {
		h.writeWebhookError(w, err)
		return
	}

	// Acknowledge the event quickly. Processing is synchronous but bounded;
	// the design keeps the handler compatible with future async processing.
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// writeWebhookError maps domain errors to appropriate HTTP status codes
// without exposing internal details, signatures, or payload contents.
func (h *WebhookHandler) writeWebhookError(w http.ResponseWriter, err error) {
	code := status.Code(err)

	switch code {
	case codes.InvalidArgument:
		http.Error(w, "invalid webhook request", http.StatusBadRequest)
	case codes.NotFound:
		http.Error(w, "webhook target not found", http.StatusNotFound)
	case codes.AlreadyExists:
		// Duplicate event: acknowledge as success so the provider stops retrying.
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	case codes.FailedPrecondition:
		http.Error(w, "webhook not permitted", http.StatusBadRequest)
	default:
		h.logger.Error().Err(err).Msg("webhook processing failed")
		http.Error(w, "webhook processing failed", http.StatusInternalServerError)
	}
}