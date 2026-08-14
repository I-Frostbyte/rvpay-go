package payments

import (
	"context"
	"errors"
	"strings"

	transactionsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/repo"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/sqlc"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Impl owns the payment-provider backend contract for the GHL Custom Payment
// Provider integration. It implements payment verification (the verify query
// operation) and payment webhook business processing. Payment-domain
// decisions live here, not in Clients; the GHL-facing transport adapters in
// Clients delegate to this service via gRPC.
type Impl struct {
	depositRepo repo.DepositRepo
	logger      zerolog.Logger

	transactionsgrpc.UnimplementedPaymentServiceServer
}

// NewPaymentService creates a new payment-provider service.
func NewPaymentService(
	depositRepo repo.DepositRepo,
	logger zerolog.Logger,
) *Impl {
	return &Impl{
		depositRepo: depositRepo,
		logger:      logger,
	}
}

// VerifyPayment verifies whether a referenced payment has succeeded. It
// looks up the deposit by its GoHighLevel transaction identifier and
// interprets its lifecycle state. Only the payment-domain status decision is
// made here; the transport adapter never interprets transaction state.
func (s *Impl) VerifyPayment(ctx context.Context, req *transactionsgrpc.VerifyPaymentRequest) (*transactionsgrpc.VerifyPaymentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "verify payment request is required")
	}

	ghlTransactionID := strings.TrimSpace(req.GetGhlTransactionId())
	if ghlTransactionID == "" {
		return nil, status.Error(codes.InvalidArgument, "ghl_transaction_id is required")
	}

	deposit, err := s.depositRepo.GetByGHLTransactionID(ctx, ghlTransactionID)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			return nil, status.Error(codes.NotFound, "transaction not found")
		default:
			s.logger.Error().Err(err).Str("ghl_transaction_id", ghlTransactionID).Msg("could not verify payment")
			return nil, status.Error(codes.Internal, "could not verify payment")
		}
	}

	// Interpret the deposit lifecycle state. Only the payment domain decides
	// what COMPLETED/FAILED/INITIATED/PROCESSING mean for the provider contract.
	switch deposit.Status {
	case sqlc.DepositStatusCOMPLETED:
		return &transactionsgrpc.VerifyPaymentResponse{Success: true}, nil
	case sqlc.DepositStatusFAILED:
		return &transactionsgrpc.VerifyPaymentResponse{Failed: true}, nil
	default:
		// INITIATED and PROCESSING are pending.
		return &transactionsgrpc.VerifyPaymentResponse{}, nil
	}
}

// ProcessPaymentWebhook processes a payment-provider webhook event. It
// correlates the HighLevel transaction/charge with an RVPay deposit and
// records the GHL reference on the deposit. Only one-time payment events
// relevant to RVPay are processed (payment.captured); subscription events are
// not supported and are acknowledged safely. Webhook idempotency is enforced
// by the transport adapter (Clients) via the webhook_events table; recording
// the GHL reference is naturally idempotent.
func (s *Impl) ProcessPaymentWebhook(ctx context.Context, req *transactionsgrpc.ProcessPaymentWebhookRequest) (*transactionsgrpc.ProcessPaymentWebhookResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "payment webhook request is required")
	}

	eventType := strings.TrimSpace(req.GetEventType())
	if eventType == "" {
		return nil, status.Error(codes.InvalidArgument, "event_type is required")
	}

	// Only payment.captured is relevant to the current one-time payment flow.
	// Unknown event types are acknowledged safely without processing.
	if eventType != "payment.captured" {
		s.logger.Info().Str("event_type", eventType).Msg("Unhandled payment webhook event type")
		return &transactionsgrpc.ProcessPaymentWebhookResponse{}, nil
	}

	if strings.TrimSpace(req.GetTransactionId()) == "" {
		s.logger.Warn().Str("event_type", eventType).Msg("payment.captured event missing transaction_id")
		return &transactionsgrpc.ProcessPaymentWebhookResponse{}, nil
	}

	// Correlate the HighLevel transaction with an RVPay deposit. If the
	// deposit is not found, acknowledge safely; the event is already recorded
	// for idempotency by the transport adapter and a later reconciliation can
	// resolve it.
	deposit, err := s.depositRepo.GetByGHLTransactionID(ctx, req.GetTransactionId())
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			s.logger.Warn().Str("transaction_id", req.GetTransactionId()).Msg("payment.captured event references unknown transaction")
			return &transactionsgrpc.ProcessPaymentWebhookResponse{}, nil
		}
		s.logger.Error().Err(err).Str("transaction_id", req.GetTransactionId()).Msg("could not correlate payment.captured event")
		return nil, status.Error(codes.Internal, "could not correlate payment event")
	}

	// Record the GHL charge reference on the deposit. The charge ID is the
	// provider's charge reference; it is correlated with the RVPay deposit so
	// charge-scoped lookups (and future reconciliation) can resolve it.
	if req.GetChargeId() != "" {
		_, err = s.depositRepo.UpdateGHLReference(ctx, deposit.ID, req.GetTransactionId(), req.GetChargeId())
		if err != nil {
			s.logger.Error().Err(err).Str("deposit_id", deposit.ID.String()).Msg("could not record GHL reference on deposit")
			return nil, status.Error(codes.Internal, "could not record payment reference")
		}
	}

	s.logger.Info().
		Str("deposit_id", deposit.ID.String()).
		Str("transaction_id", req.GetTransactionId()).
		Str("charge_id", req.GetChargeId()).
		Msg("payment.captured event processed")

	return &transactionsgrpc.ProcessPaymentWebhookResponse{}, nil
}
