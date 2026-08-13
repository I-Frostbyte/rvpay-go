package deposits

import (
	"context"
	"errors"
	"strings"

	commongrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/commongrpc"
	transactionsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/repo"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Impl implements the DepositService gRPC server.
type Impl struct {
	depositRepo  repo.DepositRepo
	customerRepo repo.CustomerRepo
	logger       zerolog.Logger

	transactionsgrpc.UnimplementedDepositServiceServer
}

// NewDepositService creates a new deposit service.
func NewDepositService(
	depositRepo repo.DepositRepo,
	customerRepo repo.CustomerRepo,
	logger zerolog.Logger,
) *Impl {
	return &Impl{
		depositRepo:  depositRepo,
		customerRepo: customerRepo,
		logger:       logger,
	}
}

// InitiateDeposit initiates a customer deposit.
func (s *Impl) InitiateDeposit(ctx context.Context, req *transactionsgrpc.CreateDepositRequest) (*transactionsgrpc.CreateDepositResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "deposit request is required")
	}

	clientID, err := uuid.Parse(req.GetClientId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "client_id must be a valid UUID")
	}

	customerID, err := uuid.Parse(req.GetCustomerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "customer_id must be a valid UUID")
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

	paymentType, err := grpcPaymentTypeToSqlc(req.GetPaymentType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	phoneNumber := strings.TrimSpace(req.GetPayerPhoneNumber())
	if phoneNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "payer_phone_number is required")
	}

	provider, err := grpcProviderToSqlc(req.GetProvider())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Validate that the customer belongs to the client/merchant context
	// before associating the deposit. This preserves the tenant boundary.
	customer, err := s.customerRepo.GetByClientAndMerchantAndPhone(ctx, clientID, merchantID, phoneNumber)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			return nil, status.Error(codes.NotFound, "customer not found for the given client, merchant, and phone number")
		default:
			s.logger.Error().Err(err).Str("customer_id", customerID.String()).Msg("could not validate customer for deposit")
			return nil, status.Error(codes.Internal, "could not validate customer for deposit")
		}
	}

	// A newly initiated deposit begins in the INITIATED lifecycle state.
	// An idempotency key is generated server-side for duplicate detection.
	deposit, err := s.depositRepo.Create(ctx, clientID, customer.ID, merchantID, amount, currency, paymentType, phoneNumber, provider, sqlc.DepositStatusINITIATED, uuid.New())
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrDuplicate):
			return nil, status.Error(codes.AlreadyExists, "deposit already exists")
		case errors.Is(err, repo.ErrConstraint):
			return nil, status.Error(codes.NotFound, "referenced merchant or customer not found")
		default:
			s.logger.Error().Err(err).Str("client_id", clientID.String()).Msg("could not create deposit")
			return nil, status.Error(codes.Internal, "could not create deposit")
		}
	}

	s.logger.Info().Str("deposit_id", deposit.ID.String()).Str("merchant_id", merchantID.String()).Msg("deposit initiated")

	return &transactionsgrpc.CreateDepositResponse{
		Deposit: depositToProto(deposit),
	}, nil
}

// GetDepositByGHLTransactionID fetches a deposit by its GoHighLevel
// transaction identifier. This is used by the GHL Custom Payment Provider
// query endpoint to correlate a HighLevel transaction with an RVPay deposit.
func (s *Impl) GetDepositByGHLTransactionID(ctx context.Context, req *transactionsgrpc.GetDepositByGHLTransactionIDRequest) (*transactionsgrpc.GetDepositByGHLTransactionIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "deposit request is required")
	}

	ghlTransactionID := strings.TrimSpace(req.GetGhlTransactionId())
	if ghlTransactionID == "" {
		return nil, status.Error(codes.InvalidArgument, "ghl_transaction_id is required")
	}

	deposit, err := s.depositRepo.GetByGHLTransactionID(ctx, ghlTransactionID)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			return nil, status.Error(codes.NotFound, "deposit not found")
		default:
			s.logger.Error().Err(err).Str("ghl_transaction_id", ghlTransactionID).Msg("could not get deposit by GHL transaction id")
			return nil, status.Error(codes.Internal, "could not get deposit")
		}
	}

	return &transactionsgrpc.GetDepositByGHLTransactionIDResponse{
		Deposit: depositToProto(deposit),
	}, nil
}

// GetDeposit fetches a deposit by id.
func (s *Impl) GetDeposit(ctx context.Context, req *transactionsgrpc.GetDepositRequest) (*transactionsgrpc.GetDepositResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "deposit request is required")
	}

	depositID, err := uuid.Parse(req.GetDepositId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "deposit_id must be a valid UUID")
	}

	deposit, err := s.depositRepo.GetByID(ctx, depositID)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			return nil, status.Error(codes.NotFound, "deposit not found")
		default:
			s.logger.Error().Err(err).Str("deposit_id", depositID.String()).Msg("could not get deposit")
			return nil, status.Error(codes.Internal, "could not get deposit")
		}
	}

	return &transactionsgrpc.GetDepositResponse{
		Deposit: depositToProto(deposit),
	}, nil
}

// validateAmount validates and converts a protobuf Money amount to pgtype.Numeric.
func validateAmount(money *commongrpc.Money) (pgtype.Numeric, error) {
	if money == nil {
		return pgtype.Numeric{}, status.Error(codes.InvalidArgument, "amount is required")
	}

	var amount pgtype.Numeric
	if err := amount.Scan(money.GetAmount()); err != nil {
		return pgtype.Numeric{}, status.Errorf(codes.InvalidArgument, "invalid deposit amount: %v", err)
	}

	f, err := amount.Float64Value()
	if err != nil {
		return pgtype.Numeric{}, status.Errorf(codes.InvalidArgument, "invalid deposit amount: %v", err)
	}
	if !f.Valid || f.Float64 <= 0 {
		return pgtype.Numeric{}, status.Error(codes.InvalidArgument, "deposit amount must be greater than zero")
	}

	return amount, nil
}

func grpcPaymentTypeToSqlc(paymentType commongrpc.PaymentType) (sqlc.PaymentType, error) {
	switch paymentType {
	case commongrpc.PaymentType_PAYMENT_TYPE_MMO:
		return sqlc.PaymentTypeMMO, nil
	case commongrpc.PaymentType_PAYMENT_TYPE_CREDIT_CARD:
		return sqlc.PaymentTypeCREDITCARD, nil
	default:
		return "", status.Errorf(codes.InvalidArgument, "unsupported payment type: %s", paymentType)
	}
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
