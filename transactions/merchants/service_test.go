package merchants

import (
	"context"
	"errors"
	"testing"

	commongrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/commongrpc"
	transactionsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/repo"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/repo/mocks"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/sqlc"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateMerchant(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  *transactionsgrpc.CreateMerchantRequest
		code codes.Code
	}{
		{name: "missing request", code: codes.InvalidArgument},
		{name: "empty name", req: &transactionsgrpc.CreateMerchantRequest{Name: "", Slug: "test"}, code: codes.InvalidArgument},
		{name: "empty slug", req: &transactionsgrpc.CreateMerchantRequest{Name: "Test Merchant", Slug: ""}, code: codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			merchantRepo := mocks.NewMockMerchantRepo(ctrl)
			service := NewMerchantService(merchantRepo, zerolog.Nop())

			_, err := service.CreateMerchant(context.Background(), tt.req)
			if got := status.Code(err); got != tt.code {
				t.Fatalf("status code = %s, want %s", got, tt.code)
			}
		})
	}
}

func TestCreateMerchantSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	merchantRepo := mocks.NewMockMerchantRepo(ctrl)
	service := NewMerchantService(merchantRepo, zerolog.Nop())

	merchantRepo.EXPECT().Create(gomock.Any(), "Test Merchant", "test-merchant", sqlc.MerchantStatusONBOARDED).
		Return(sqlc.Merchant{
			ID:   uuid.New(),
			Name: "Test Merchant",
			Slug: "test-merchant",
		}, nil)

	resp, err := service.CreateMerchant(context.Background(), &transactionsgrpc.CreateMerchantRequest{
		Name: "Test Merchant",
		Slug: "test-merchant",
	})
	if err != nil {
		t.Fatalf("CreateMerchant failed: %v", err)
	}
	if resp.Merchant.Name != "Test Merchant" {
		t.Fatalf("merchant name = %s, want Test Merchant", resp.Merchant.Name)
	}
}

func TestCreateMerchantDuplicate(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	merchantRepo := mocks.NewMockMerchantRepo(ctrl)
	service := NewMerchantService(merchantRepo, zerolog.Nop())

	merchantRepo.EXPECT().Create(gomock.Any(), "Test Merchant", "test-merchant", sqlc.MerchantStatusONBOARDED).
		Return(sqlc.Merchant{}, repo.ErrDuplicate)

	_, err := service.CreateMerchant(context.Background(), &transactionsgrpc.CreateMerchantRequest{
		Name: "Test Merchant",
		Slug: "test-merchant",
	})
	if got := status.Code(err); got != codes.AlreadyExists {
		t.Fatalf("status code = %s, want %s", got, codes.AlreadyExists)
	}
}

func TestGetMerchant(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	merchantRepo := mocks.NewMockMerchantRepo(ctrl)
	service := NewMerchantService(merchantRepo, zerolog.Nop())

	merchantID := uuid.New()
	merchantRepo.EXPECT().GetByID(gomock.Any(), merchantID).
		Return(sqlc.Merchant{
			ID:   merchantID,
			Name: "Test Merchant",
			Slug: "test-merchant",
		}, nil)

	resp, err := service.GetMerchant(context.Background(), &transactionsgrpc.GetMerchantRequest{
		MerchantId: merchantID.String(),
	})
	if err != nil {
		t.Fatalf("GetMerchant failed: %v", err)
	}
	if resp.Merchant.Id != merchantID.String() {
		t.Fatalf("merchant id = %s, want %s", resp.Merchant.Id, merchantID.String())
	}
}

func TestGetMerchantNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	merchantRepo := mocks.NewMockMerchantRepo(ctrl)
	service := NewMerchantService(merchantRepo, zerolog.Nop())

	merchantRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).
		Return(sqlc.Merchant{}, repo.ErrNotFound)

	_, err := service.GetMerchant(context.Background(), &transactionsgrpc.GetMerchantRequest{
		MerchantId: uuid.New().String(),
	})
	if got := status.Code(err); got != codes.NotFound {
		t.Fatalf("status code = %s, want %s", got, codes.NotFound)
	}
}

func TestGetMerchantInvalidID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	merchantRepo := mocks.NewMockMerchantRepo(ctrl)
	service := NewMerchantService(merchantRepo, zerolog.Nop())

	_, err := service.GetMerchant(context.Background(), &transactionsgrpc.GetMerchantRequest{
		MerchantId: "not-a-uuid",
	})
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Fatalf("status code = %s, want %s", got, codes.InvalidArgument)
	}
}

func TestListMerchants(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	merchantRepo := mocks.NewMockMerchantRepo(ctrl)
	service := NewMerchantService(merchantRepo, zerolog.Nop())

	// With no page request, the default page size (20) and offset 0 are used.
	merchantRepo.EXPECT().List(gomock.Any(), int32(20), int32(0)).
		Return([]sqlc.Merchant{
			{ID: uuid.New(), Name: "Merchant 1", Slug: "merchant-1"},
			{ID: uuid.New(), Name: "Merchant 2", Slug: "merchant-2"},
		}, nil)
	merchantRepo.EXPECT().Count(gomock.Any()).
		Return(int64(2), nil)

	resp, err := service.ListMerchants(context.Background(), &transactionsgrpc.ListMerchantsRequest{})
	if err != nil {
		t.Fatalf("ListMerchants failed: %v", err)
	}
	if len(resp.Merchants) != 2 {
		t.Fatalf("merchants count = %d, want 2", len(resp.Merchants))
	}
	if resp.Page.TotalCount != 2 {
		t.Fatalf("total count = %d, want 2", resp.Page.TotalCount)
	}
	if resp.Page.NextPageToken != "" {
		t.Fatalf("next page token = %q, want empty (no more pages)", resp.Page.NextPageToken)
	}
}

func TestListMerchantsHonorsPageSizeAndCap(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	merchantRepo := mocks.NewMockMerchantRepo(ctrl)
	service := NewMerchantService(merchantRepo, zerolog.Nop())

	// A requested page size above the max (100) must be capped to 100.
	merchantRepo.EXPECT().List(gomock.Any(), int32(100), int32(0)).
		Return([]sqlc.Merchant{
			{ID: uuid.New(), Name: "Merchant 1", Slug: "merchant-1"},
		}, nil)
	merchantRepo.EXPECT().Count(gomock.Any()).
		Return(int64(150), nil)

	resp, err := service.ListMerchants(context.Background(), &transactionsgrpc.ListMerchantsRequest{
		Page: &commongrpc.PaginationRequest{PageSize: 1000},
	})
	if err != nil {
		t.Fatalf("ListMerchants failed: %v", err)
	}
	if len(resp.Merchants) != 1 {
		t.Fatalf("merchants count = %d, want 1", len(resp.Merchants))
	}
	if resp.Page.TotalCount != 150 {
		t.Fatalf("total count = %d, want 150", resp.Page.TotalCount)
	}
	// offset(0) + returned(1) < total(150) => a next page exists.
	if resp.Page.NextPageToken == "" {
		t.Fatal("next page token should be set when more pages exist")
	}
}

func TestListMerchantsRepositoryError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	merchantRepo := mocks.NewMockMerchantRepo(ctrl)
	service := NewMerchantService(merchantRepo, zerolog.Nop())

	merchantRepo.EXPECT().List(gomock.Any(), int32(20), int32(0)).
		Return(nil, errors.New("database down"))

	_, err := service.ListMerchants(context.Background(), &transactionsgrpc.ListMerchantsRequest{})
	if got := status.Code(err); got != codes.Internal {
		t.Fatalf("status code = %s, want %s", got, codes.Internal)
	}
}
