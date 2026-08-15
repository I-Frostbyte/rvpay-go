package payments

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"strings"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	transactionsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// QueryOperation is the HighLevel payment query operation type.
type QueryOperation string

const (
	// QueryOperationVerify is the verify operation. HighLevel sends this to
	// determine whether a referenced transaction has succeeded.
	QueryOperationVerify QueryOperation = "verify"
)

// QueryRequest is the HighLevel payment query request.
type QueryRequest struct {
	Type           string `json:"type"`
	TransactionID  string `json:"transactionId"`
	APIKey         string `json:"apiKey"`
	ChargeID       string `json:"chargeId"`
	SubscriptionID string `json:"subscriptionId"`
}

// QueryResponse is the HighLevel payment query response.
type QueryResponse struct {
	Success bool `json:"success,omitempty"`
	Failed  bool `json:"failed,omitempty"`
}

// WebhookEvent is the HighLevel payment-provider webhook event.
type WebhookEvent struct {
	EventType     string                 `json:"eventType"`
	EventID       string                 `json:"eventId"`
	LocationID    string                 `json:"locationId"`
	ChargeID      string                 `json:"chargeId"`
	TransactionID string                 `json:"transactionId"`
	APIKey        string                 `json:"apiKey"`
	Data          map[string]interface{} `json:"data"`
}

// Service is the GHL Custom Payment Provider transport adapter. It owns the
// GHL-facing request parsing, credential validation, provider configuration
// lookup, and webhook idempotency. It does NOT own payment-domain business
// logic: payment verification, transaction lookup, payment state
// interpretation, and payment webhook business processing are delegated to
// the Transactions service via gRPC.
type Service struct {
	configRepo         repo.PaymentProviderConfigRepo
	integrationRepo    repo.IntegrationRepo
	webhookEventRepo   repo.WebhookEventRepo
	transactionsClient transactionsgrpc.PaymentServiceClient
	logger             zerolog.Logger
}

// NewService creates a new GHL Custom Payment Provider transport adapter.
// transactionsClient is the gRPC client used to delegate payment-domain
// decisions (verification and webhook business processing) to the
// Transactions service.
func NewService(
	configRepo repo.PaymentProviderConfigRepo,
	integrationRepo repo.IntegrationRepo,
	webhookEventRepo repo.WebhookEventRepo,
	transactionsClient transactionsgrpc.PaymentServiceClient,
	logger zerolog.Logger,
) *Service {
	return &Service{
		configRepo:         configRepo,
		integrationRepo:    integrationRepo,
		webhookEventRepo:   webhookEventRepo,
		transactionsClient: transactionsClient,
		logger:             logger,
	}
}

// HandleQuery processes a HighLevel payment query request. It validates the
// provider API key, inspects the operation type, and for the verify operation
// delegates the payment-domain decision to the Transactions service. It
// returns the HighLevel contract response and never leaks internal RVPay
// transaction objects or database models.
func (s *Service) HandleQuery(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "query request is required")
	}

	if strings.TrimSpace(req.APIKey) == "" {
		return nil, ErrMissingAPIKey
	}

	// Resolve the provider configuration by location. The provider API key is
	// stored per-integration and is distinct from the OAuth client secret and
	// the pawaPay API key. We look up the config by the API key to avoid
	// trusting an arbitrary locationId from the request.
	config, err := s.findConfigByAPIKey(ctx, req.APIKey)
	if err != nil {
		return nil, err
	}

	switch QueryOperation(strings.TrimSpace(req.Type)) {
	case QueryOperationVerify:
		return s.handleVerify(ctx, config, req)
	default:
		return nil, ErrUnsupportedType
	}
}

// handleVerify processes the verify operation. It delegates the payment-domain
// decision (transaction lookup, state interpretation) to the Transactions
// service and maps the result to the HighLevel contract response.
func (s *Service) handleVerify(ctx context.Context, config sqlc.PaymentProviderConfig, req *QueryRequest) (*QueryResponse, error) {
	if strings.TrimSpace(req.TransactionID) == "" && strings.TrimSpace(req.ChargeID) == "" {
		return nil, ErrMissingTransactionID
	}

	// Delegate the payment-domain decision to the Transactions service. The
	// transport adapter never interprets transaction state. The charge ID is
	// passed as a fallback lookup key; the subscription ID is passed so the
	// Transactions service can reject subscription-scoped verifications.
	resp, err := s.transactionsClient.VerifyPayment(ctx, &transactionsgrpc.VerifyPaymentRequest{
		GhlTransactionId: req.TransactionID,
		GhlChargeId:      req.ChargeID,
		SubscriptionId:   req.SubscriptionID,
	})
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			return nil, ErrTransactionNotFound
		case codes.FailedPrecondition:
			// Subscription-scoped verification is not supported by RVPay.
			return nil, status.Error(codes.FailedPrecondition, "subscription payments are not supported")
		case codes.InvalidArgument:
			return nil, status.Error(codes.InvalidArgument, "invalid payment verification request")
		default:
			s.logger.Error().Err(err).Str("transaction_id", req.TransactionID).Msg("could not verify transaction with Transactions service")
			return nil, status.Error(codes.Internal, "could not verify transaction")
		}
	}

	return &QueryResponse{
		Success: resp.GetSuccess(),
		Failed:  resp.GetFailed(),
	}, nil
}

// HandleWebhook processes a HighLevel payment-provider webhook event. It
// authenticates the request using the provider API key, enforces idempotency
// via the webhook_events table unique constraint, and delegates payment-domain
// business processing to the Transactions service. The provider API key is
// never logged or returned in error messages.
func (s *Service) HandleWebhook(ctx context.Context, body []byte) error {
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		s.logger.Error().Err(err).Msg("failed to parse payment webhook payload")
		return ErrInvalidPayload
	}

	if strings.TrimSpace(event.EventID) == "" {
		return ErrInvalidPayload
	}

	// Resolve the integration by location ID. The payment-provider webhook is
	// location-scoped; the location ID maps to the integration's payment
	// provider configuration.
	config, err := s.configRepo.GetByLocationID(ctx, event.LocationID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrProviderConfigNotFound
		}
		return translateError(err)
	}

	// Authenticate the webhook request using the provider API key. The key is
	// validated against the stored config using constant-time comparison to
	// avoid timing side-channels. The key is never logged or returned in
	// errors.
	if strings.TrimSpace(event.APIKey) == "" {
		return ErrMissingAPIKey
	}
	if subtle.ConstantTimeCompare([]byte(config.ProviderApiKey), []byte(event.APIKey)) != 1 {
		return ErrInvalidAPIKey
	}

	integration, err := s.integrationRepo.GetByID(ctx, config.IntegrationID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrIntegrationNotFound
		}
		return translateError(err)
	}

	// Idempotency: record the event atomically. The unique constraint on
	// (integration_id, provider_event_id) plus ON CONFLICT DO NOTHING makes
	// duplicate deliveries race-safe.
	payload, err := json.Marshal(event.Data)
	if err != nil {
		s.logger.Error().Err(err).Str("event_id", event.EventID).Msg("payment webhook payload marshal failed")
		return ErrInvalidPayload
	}

	_, err = s.webhookEventRepo.Create(ctx, integration.ID, event.EventID, event.EventType, payload)
	if err == repo.ErrDuplicate {
		s.logger.Info().Str("event_id", event.EventID).Str("integration_id", integration.ID.String()).Msg("Duplicate payment webhook event ignored")
		return ErrDuplicateEvent
	}
	if err != nil {
		return translateError(err)
	}

	// Delegate payment-domain business processing to the Transactions service.
	// The transport adapter does not interpret event semantics.
	_, err = s.transactionsClient.ProcessPaymentWebhook(ctx, &transactionsgrpc.ProcessPaymentWebhookRequest{
		EventType:     event.EventType,
		TransactionId: event.TransactionID,
		ChargeId:      event.ChargeID,
		LocationId:    event.LocationID,
	})
	if err != nil {
		s.logger.Error().Err(err).Str("event_id", event.EventID).Msg("could not process payment webhook with Transactions service")
		return status.Error(codes.Internal, "could not process payment webhook")
	}

	return nil
}

// findConfigByAPIKey resolves the payment provider configuration by matching
// the provider API key. It uses constant-time comparison to avoid timing
// side-channels and never logs the API key.
func (s *Service) findConfigByAPIKey(ctx context.Context, apiKey string) (sqlc.PaymentProviderConfig, error) {
	config, err := s.configRepo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return sqlc.PaymentProviderConfig{}, ErrInvalidAPIKey
		}
		return sqlc.PaymentProviderConfig{}, translateError(err)
	}

	// Constant-time comparison as a defense-in-depth check. The repo lookup
	// already matches the key; this guards against any future change.
	if subtle.ConstantTimeCompare([]byte(config.ProviderApiKey), []byte(apiKey)) != 1 {
		return sqlc.PaymentProviderConfig{}, ErrInvalidAPIKey
	}

	return config, nil
}
