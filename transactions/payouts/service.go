package payouts

import (
	"context"
	"errors"
	"strings"

	"github.com/I-Frostbyte/rvpay-go/transactions/db/repo"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/sqlc"
	commongrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/commongrpc"
	transactionsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Impl implements the PayoutService gRPC server.
type Impl struct {
	payoutRepo repo.PayoutRepo
	logger     zerolog.Logger

	transactionsgrpc.UnimplementedPayoutServiceServer
}

// NewPayoutService creates a new payout service.
func NewPayoutService(
	payoutRepo repo.PayoutRepo,
	logger zerolog.Logger,
) *Impl {
	return &Impl{
		payoutRepo: payoutRepo,
		logger:     logger,
	}
}

// RequestPayout requests an outbound settlement.
func (s *Impl) RequestPayout(ctx context.Context, req *transactionsgrpc.CreatePayoutRequest) (*transactionsgrpc.CreatePayoutResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "payout request is required")
	}

	clientID, err := uuid.Parse(req.GetClientId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "client_id must be a valid UUID")
	}

	merchantID, err := uuid.Parse(req.GetMerchantId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "merchant_id must be a valid UUID")
	}

	amount, err := validateAmount(req.GetAmount())
	if err != nil {
		return nil, err
	}

	currency := strings.ToUpper(strings.TrimSpace(req.GetAmount().GetCurrency()))
	if currency == "" {
		return nil, status.Error(codes.InvalidArgument, "currency is required")
	}

	provider, err := grpcProviderToSqlc(req.GetProvider())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	destinationReference := strings.TrimSpace(req.GetDestinationReference())
	if destinationReference == "" {
		return nil, status.Error(codes.InvalidArgument, "destination_reference is required")
	}

	// A newly requested payout begins in the REQUESTED lifecycle state.
	// An idempotency key is generated server-side for duplicate detection.
	// Merchant existence is enforced by the database foreign key.
	payout, err := s.payoutRepo.Create(ctx, clientID, merchantID, amount, currency, provider, destinationReference, sqlc.PayoutStatusREQUESTED, uuid.New())
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrDuplicate):
			return nil, status.Error(codes.AlreadyExists, "payout already exists")
		case errors.Is(err, repo.ErrConstraint):
			return nil, status.Error(codes.NotFound, "referenced merchant not found")
		default:
			s.logger.Error().Err(err).Str("client_id", clientID.String()).Msg("could not create payout")
			return nil, status.Error(codes.Internal, "could not create payout")
		}
	}

	s.logger.Info().Str("payout_id", payout.ID.String()).Str("merchant_id", merchantID.String()).Msg("payout requested")

	return &transactionsgrpc.CreatePayoutResponse{
		Payout: payoutToProto(payout),
	}, nil
}

// GetPayout fetches a payout by id.
func (s *Impl) GetPayout(ctx context.Context, req *transactionsgrpc.GetPayoutRequest) (*transactionsgrpc.GetPayoutResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "payout request is required")
	}

	payoutID, err := uuid.Parse(req.GetPayoutId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "payout_id must be a valid UUID")
	}

	payout, err := s.payoutRepo.GetByID(ctx, payoutID)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			return nil, status.Error(codes.NotFound, "payout not found")
		default:
			s.logger.Error().Err(err).Str("payout_id", payoutID.String()).Msg("could not get payout")
			return nil, status.Error(codes.Internal, "could not get payout")
		}
	}

	return &transactionsgrpc.GetPayoutResponse{
		Payout: payoutToProto(payout),
	}, nil
}

// validateAmount validates and converts a protobuf Money amount to pgtype.Numeric.
func validateAmount(money *commongrpc.Money) (pgtype.Numeric, error) {
	if money == nil {
		return pgtype.Numeric{}, status.Error(codes.InvalidArgument, "amount is required")
	}

	var amount pgtype.Numeric
	if err := amount.Scan(money.GetAmount()); err != nil {
		return pgtype.Numeric{}, status.Errorf(codes.InvalidArgument, "invalid payout amount: %v", err)
	}

	f, err := amount.Float64Value()
	if err != nil {
		return pgtype.Numeric{}, status.Errorf(codes.InvalidArgument, "invalid payout amount: %v", err)
	}
	if !f.Valid || f.Float64 <= 0 {
		return pgtype.Numeric{}, status.Error(codes.InvalidArgument, "payout amount must be greater than zero")
	}

	return amount, nil
}

func grpcProviderToSqlc(provider commongrpc.Provider) (sqlc.PaymentProvider, error) {
	switch provider {
	case commongrpc.Provider_PROVIDER_MTN_MOMO:
		return sqlc.PaymentProviderMTNMOMO, nil
	case commongrpc.Provider_PROVIDER_ORANGE_MOMO:
		return sqlc.PaymentProviderORANGEMOMO, nil
	default:
		return "", status.Errorf(codes.InvalidArgument, "unsupported provider: %s", provider)
	}
}