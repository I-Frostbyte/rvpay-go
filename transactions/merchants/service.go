package merchants

import (
	"context"
	"errors"
	"strings"

	commongrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/commongrpc"
	transactionsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/repo"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/sqlc"
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

// defaultPageSize is the default number of merchants returned per page when
// the client does not specify a page size.
const defaultPageSize = int32(20)

// maxPageSize caps the number of merchants a client may request per page so
// that a single request cannot force an unbounded result set.
const maxPageSize = int32(100)

// ListMerchants lists merchants with pagination.
func (s *Impl) ListMerchants(ctx context.Context, req *transactionsgrpc.ListMerchantsRequest) (*transactionsgrpc.ListMerchantsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "merchant request is required")
	}

	pageSize := defaultPageSize
	offset := int32(0)
	if req.GetPage() != nil {
		if req.GetPage().GetPageSize() > 0 {
			pageSize = req.GetPage().GetPageSize()
		}
		if pageSize > maxPageSize {
			pageSize = maxPageSize
		}
		if req.GetPage().GetPageToken() != "" {
			// The page token is an opaque cursor; for the initial offset-based
			// implementation it advances by one page of the requested size.
			offset = pageSize
		}
	}

	merchants, err := s.merchantRepo.List(ctx, pageSize, offset)
	if err != nil {
		s.logger.Error().Err(err).Msg("could not list merchants")
		return nil, status.Error(codes.Internal, "could not list merchants")
	}

	total, err := s.merchantRepo.Count(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("could not count merchants")
		return nil, status.Error(codes.Internal, "could not count merchants")
	}

	protoMerchants := make([]*transactionsgrpc.Merchant, 0, len(merchants))
	for _, merchant := range merchants {
		protoMerchants = append(protoMerchants, merchantToProto(merchant))
	}

	nextPageToken := ""
	if int64(offset)+int64(len(merchants)) < total {
		nextPageToken = "next"
	}

	response := &transactionsgrpc.ListMerchantsResponse{
		Merchants: protoMerchants,
		Page: &commongrpc.PaginationResponse{
			NextPageToken: nextPageToken,
			TotalCount:    total,
		},
	}

	return response, nil
}
