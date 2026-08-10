package merchants

import (
	"context"
	"errors"
	"strings"

	"github.com/I-Frostbyte/rvpay-go/transactions/db/repo"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/sqlc"
	commongrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/commongrpc"
	transactionsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Impl implements the MerchantService gRPC server.
type Impl struct {
	merchantRepo repo.MerchantRepo
	logger       zerolog.Logger

	transactionsgrpc.UnimplementedMerchantServiceServer
}

// NewMerchantService creates a new merchant service.
func NewMerchantService(
	merchantRepo repo.MerchantRepo,
	logger zerolog.Logger,
) *Impl {
	return &Impl{
		merchantRepo: merchantRepo,
		logger:       logger,
	}
}

// CreateMerchant registers a new merchant.
func (s *Impl) CreateMerchant(ctx context.Context, req *transactionsgrpc.CreateMerchantRequest) (*transactionsgrpc.CreateMerchantResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "merchant request is required")
	}

	name := strings.TrimSpace(req.GetName())
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant name is required")
	}

	slug := strings.TrimSpace(req.GetSlug())
	if slug == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant slug is required")
	}

	// A newly registered merchant begins in the ONBOARDED lifecycle state.
	merchant, err := s.merchantRepo.Create(ctx, name, slug, sqlc.MerchantStatusONBOARDED)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrDuplicate):
			return nil, status.Error(codes.AlreadyExists, "merchant already exists")
		default:
			s.logger.Error().Err(err).Str("name", name).Str("slug", slug).Msg("could not create merchant")
			return nil, status.Error(codes.Internal, "could not create merchant")
		}
	}

	s.logger.Info().Str("merchant_id", merchant.ID.String()).Str("slug", merchant.Slug).Msg("merchant created")

	return &transactionsgrpc.CreateMerchantResponse{
		Merchant: merchantToProto(merchant),
	}, nil
}

// GetMerchant fetches a merchant by id.
func (s *Impl) GetMerchant(ctx context.Context, req *transactionsgrpc.GetMerchantRequest) (*transactionsgrpc.GetMerchantResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "merchant request is required")
	}

	merchantID, err := uuid.Parse(req.GetMerchantId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "merchant_id must be a valid UUID")
	}

	merchant, err := s.merchantRepo.GetByID(ctx, merchantID)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			return nil, status.Error(codes.NotFound, "merchant not found")
		default:
			s.logger.Error().Err(err).Str("merchant_id", merchantID.String()).Msg("could not get merchant")
			return nil, status.Error(codes.Internal, "could not get merchant")
		}
	}

	return &transactionsgrpc.GetMerchantResponse{
		Merchant: merchantToProto(merchant),
	}, nil
}

// ListMerchants lists merchants.
func (s *Impl) ListMerchants(ctx context.Context, req *transactionsgrpc.ListMerchantsRequest) (*transactionsgrpc.ListMerchantsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "merchant request is required")
	}

	merchants, err := s.merchantRepo.List(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("could not list merchants")
		return nil, status.Error(codes.Internal, "could not list merchants")
	}

	protoMerchants := make([]*transactionsgrpc.Merchant, 0, len(merchants))
	for _, merchant := range merchants {
		protoMerchants = append(protoMerchants, merchantToProto(merchant))
	}

	// Pagination is deferred at the persistence layer; the full result set is
	// returned with an empty next-page token and the total count.
	response := &transactionsgrpc.ListMerchantsResponse{
		Merchants: protoMerchants,
		Page: &commongrpc.PaginationResponse{
			NextPageToken: "",
			TotalCount:    int64(len(protoMerchants)),
		},
	}

	return response, nil
}