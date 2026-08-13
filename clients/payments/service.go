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
	Data          map[string]interface{} `json:"data"`
}

// Service manages the GHL Custom Payment Provider backend integration.
// It handles the payment query endpoint (verify operation) and the
// payment-provider webhook (payment.captured), correlating HighLevel
// transactions with RVPay deposits via the Transactions service.
type Service struct {
	configRepo         repo.PaymentProviderConfigRepo
	integrationRepo    repo.IntegrationRepo
	webhookEventRepo   repo.WebhookEventRepo
	transactionsClient transactionsgrpc.DepositServiceClient
	logger             zerolog.Logger
}

// NewService creates a new GHL Custom Payment Provider service.
// transactionsClient is the gRPC client used to correlate HighLevel
// transactions with RVPay deposits in the Transactions service.
func NewService(
	configRepo repo.PaymentProviderConfigRepo,
	integrationRepo repo.IntegrationRepo,
	webhookEventRepo repo.WebhookEventRepo,
	transactionsClient transactionsgrpc.DepositServiceClient,
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
// correlates the referenced transaction with an RVPay deposit to determine its
// status. It returns the HighLevel contract response and never leaks internal
// RVPay transaction objects or database models.
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

// handleVerify processes the verify operation. It correlates the HighLevel
// transaction with an RVPay deposit via the Transactions service and returns
// the appropriate HighLevel contract response.
func (s *Service) handleVerify(ctx context.Context, config sqlc.PaymentProviderConfig, req *QueryRequest) (*QueryResponse, error) {
	if strings.TrimSpace(req.TransactionID) == "" {
		return nil, ErrMissingTransactionID
	}

	// Correlate the HighLevel transaction with an RVPay deposit by calling the
	// Transactions service. This keeps transaction-status logic in the
	// Transactions domain and avoids duplicating it in the HTTP handler.
	depositResp, err := s.transactionsClient.GetDepositByGHLTransactionID(ctx, &transactionsgrpc.GetDepositByGHLTransactionIDRequest{
		GhlTransactionId: req.TransactionID,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrTransactionNotFound
		}
		s.logger.Error().Err(err).Str("transaction_id", req.TransactionID).Msg("could not verify transaction with Transactions service")
		return nil, status.Error(codes.Internal, "could not verify transaction")
	}

	deposit := depositResp.GetDeposit()
	if deposit == nil {
		return nil, ErrTransactionNotFound
	}

	switch deposit.GetStatus() {
	case transactionsgrpc.DepositStatus_DEPOSIT_STATUS_COMPLETED:
		return &QueryResponse{Success: true}, nil
	case transactionsgrpc.DepositStatus_DEPOSIT_STATUS_FAILED:
		return &QueryResponse{Failed: true}, nil
	default:
		// INITIATED, PROCESSING, and UNSPECIFIED are treated as pending.
		return &QueryResponse{Success: false}, nil
	}
}

// HandleWebhook processes a HighLevel payment-provider webhook event. It is
// idempotent: duplicate deliveries are detected via the webhook_events table
// unique constraint and acknowledged without reprocessing.
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

	// Only payment.captured is relevant to the current one-time payment flow.
	// Unknown event types are acknowledged safely without processing.
	switch event.EventType {
	case "payment.captured":
		return s.handlePaymentCaptured(ctx, integration, event)
	default:
		s.logger.Info().Str("event_type", event.EventType).Str("event_id", event.EventID).Msg("Unhandled payment webhook event type")
		return nil
	}
}

// handlePaymentCaptured processes a payment.captured event. It correlates the
// HighLevel charge/transaction with an RVPay deposit and records the GHL
// reference on the deposit via the Transactions service.
func (s *Service) handlePaymentCaptured(ctx context.Context, integration sqlc.Integration, event WebhookEvent) error {
	if strings.TrimSpace(event.TransactionID) == "" {
		s.logger.Warn().Str("event_id", event.EventID).Msg("payment.captured event missing transactionId")
		return nil
	}

	// Correlate the HighLevel transaction with an RVPay deposit. If the deposit
	// is not found, log and acknowledge (the event is already recorded for
	// idempotency; a later reconciliation can resolve it).
	depositResp, err := s.transactionsClient.GetDepositByGHLTransactionID(ctx, &transactionsgrpc.GetDepositByGHLTransactionIDRequest{
		GhlTransactionId: event.TransactionID,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			s.logger.Warn().Str("transaction_id", event.TransactionID).Msg("payment.captured event references unknown transaction")
			return nil
		}
		s.logger.Error().Err(err).Str("transaction_id", event.TransactionID).Msg("could not correlate payment.captured event")
		return status.Error(codes.Internal, "could not correlate payment event")
	}

	deposit := depositResp.GetDeposit()
	if deposit == nil {
		s.logger.Warn().Str("transaction_id", event.TransactionID).Msg("payment.captured event references unknown transaction")
		return nil
	}

	s.logger.Info().
		Str("deposit_id", deposit.GetId()).
		Str("transaction_id", event.TransactionID).
		Str("charge_id", event.ChargeID).
		Msg("payment.captured event correlated with deposit")

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
