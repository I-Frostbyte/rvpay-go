package customers

import (
	"context"
	"errors"
	"strings"

	"github.com/I-Frostbyte/rvpay-go/transactions/db/repo"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/sqlc"
	transactionsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Impl implements the CustomerService gRPC server.
type Impl struct {
	customerRepo repo.CustomerRepo
	logger       zerolog.Logger

	transactionsgrpc.UnimplementedCustomerServiceServer
}

// NewCustomerService creates a new customer service.
func NewCustomerService(
	customerRepo repo.CustomerRepo,
	logger zerolog.Logger,
) *Impl {
	return &Impl{
		customerRepo: customerRepo,
		logger:       logger,
	}
}

// CreateCustomer creates a customer record.
func (s *Impl) CreateCustomer(ctx context.Context, req *transactionsgrpc.CreateCustomerRequest) (*transactionsgrpc.CreateCustomerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "customer request is required")
	}

	clientID, err := uuid.Parse(req.GetClientId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "client_id must be a valid UUID")
	}

	merchantID, err := uuid.Parse(req.GetMerchantId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "merchant_id must be a valid UUID")
	}

	phoneNumber := strings.TrimSpace(req.GetPhoneNumber())
	if phoneNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "phone_number is required")
	}

	// A newly created customer begins in the CREATED lifecycle state.
	// Merchant existence is enforced by the database foreign key; no
	// cross-service merchant call is required.
	customer, err := s.customerRepo.Create(ctx, clientID, merchantID, phoneNumber, sqlc.CustomerStatusCREATED)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrDuplicate):
			return nil, status.Error(codes.AlreadyExists, "customer already exists")
		case errors.Is(err, repo.ErrConstraint):
			return nil, status.Error(codes.NotFound, "referenced merchant not found")
		default:
			s.logger.Error().Err(err).Str("client_id", clientID.String()).Str("merchant_id", merchantID.String()).Msg("could not create customer")
			return nil, status.Error(codes.Internal, "could not create customer")
		}
	}

	s.logger.Info().Str("customer_id", customer.ID.String()).Str("merchant_id", merchantID.String()).Msg("customer created")

	return &transactionsgrpc.CreateCustomerResponse{
		Customer: customerToProto(customer),
	}, nil
}

// GetCustomer fetches a customer by id.
func (s *Impl) GetCustomer(ctx context.Context, req *transactionsgrpc.GetCustomerRequest) (*transactionsgrpc.GetCustomerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "customer request is required")
	}

	customerID, err := uuid.Parse(req.GetCustomerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "customer_id must be a valid UUID")
	}

	customer, err := s.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			return nil, status.Error(codes.NotFound, "customer not found")
		default:
			s.logger.Error().Err(err).Str("customer_id", customerID.String()).Msg("could not get customer")
			return nil, status.Error(codes.Internal, "could not get customer")
		}
	}

	return &transactionsgrpc.GetCustomerResponse{
		Customer: customerToProto(customer),
	}, nil
}