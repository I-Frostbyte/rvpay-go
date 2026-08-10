package deposits

import (
	"context"
	"errors"
	"testing"

	"github.com/I-Frostbyte/rvpay-go/transactions/db/repo"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/repo/mocks"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/sqlc"
	commongrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/commongrpc"
	transactionsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestInitiateDepositValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  *transactionsgrpc.CreateDepositRequest
		code codes.Code
	}{
		{name: "missing request", code: codes.InvalidArgument},
		{name: "invalid client id", req: &transactionsgrpc.CreateDepositRequest{ClientId: "not-a-uuid"}, code: codes.InvalidArgument},
		{name: "invalid customer id", req: &transactionsgrpc.CreateDepositRequest{ClientId: uuid.New().String(), CustomerId: "not-a-uuid"}, code: codes.InvalidArgument},
		{name: "invalid merchant id", req: &transactionsgrpc.CreateDepositRequest{ClientId: uuid.New().String(), CustomerId: uuid.New().String(), MerchantId: "not-a-uuid"}, code: codes.InvalidArgument},
		{name: "missing amount", req: &transactionsgrpc.CreateDepositRequest{ClientId: uuid.New().String(), CustomerId: uuid.New().String(), MerchantId: uuid.New().String()}, code: codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			depositRepo := mocks.NewMockDepositRepo(ctrl)
			customerRepo := mocks.NewMockCustomerRepo(ctrl)
			service := NewDepositService(depositRepo, customerRepo, zerolog.Nop())

			_, err := service.InitiateDeposit(context.Background(), tt.req)
			if got := status.Code(err); got != tt.code {
				t.Fatalf("status code = %s, want %s", got, tt.code)
			}
		})
	}
}

func TestInitiateDepositZeroAmount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	depositRepo := mocks.NewMockDepositRepo(ctrl)
	customerRepo := mocks.NewMockCustomerRepo(ctrl)
	service := NewDepositService(depositRepo, customerRepo, zerolog.Nop())

	_, err := service.InitiateDeposit(context.Background(), &transactionsgrpc.CreateDepositRequest{
		ClientId:   uuid.New().String(),
		CustomerId: uuid.New().String(),
		MerchantId: uuid.New().String(),
		Amount: &commongrpc.Money{
			Amount:   "0",
			Currency: "XAF",
		},
	})
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Fatalf("status code = %s, want %s", got, codes.InvalidArgument)
	}
}

func TestInitiateDepositCustomerNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	depositRepo := mocks.NewMockDepositRepo(ctrl)
	customerRepo := mocks.NewMockCustomerRepo(ctrl)
	service := NewDepositService(depositRepo, customerRepo, zerolog.Nop())

	customerRepo.EXPECT().GetByClientAndMerchantAndPhone(gomock.Any(), gomock.Any(), gomock.Any(), "+237600000000").
		Return(sqlc.Customer{}, repo.ErrNotFound)

	_, err := service.InitiateDeposit(context.Background(), &transactionsgrpc.CreateDepositRequest{
		ClientId:        uuid.New().String(),
		CustomerId:      uuid.New().String(),
		MerchantId:      uuid.New().String(),
		Amount:          &commongrpc.Money{Amount: "1000.00", Currency: "XAF"},
		PaymentType:     commongrpc.PaymentType_PAYMENT_TYPE_MMO,
		PayerPhoneNumber: "+237600000000",
		Provider:        commongrpc.Provider_PROVIDER_MTN_MOMO,
	})
	if got := status.Code(err); got != codes.NotFound {
		t.Fatalf("status code = %s, want %s", got, codes.NotFound)
	}
}

func TestInitiateDepositSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	depositRepo := mocks.NewMockDepositRepo(ctrl)
	customerRepo := mocks.NewMockCustomerRepo(ctrl)
	service := NewDepositService(depositRepo, customerRepo, zerolog.Nop())

	var amount pgtype.Numeric
	if err := amount.Scan("1000.00"); err != nil {
		t.Fatalf("failed to scan amount: %v", err)
	}

	customerRepo.EXPECT().GetByClientAndMerchantAndPhone(gomock.Any(), gomock.Any(), gomock.Any(), "+237600000000").
		Return(sqlc.Customer{ID: uuid.New()}, nil)

	depositRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "XAF", sqlc.PaymentTypeMMO, "+237600000000", sqlc.PaymentProviderMTNMOMO, sqlc.DepositStatusINITIATED, gomock.Any()).
		Return(sqlc.Deposit{ID: uuid.New()}, nil)

	resp, err := service.InitiateDeposit(context.Background(), &transactionsgrpc.CreateDepositRequest{
		ClientId:        uuid.New().String(),
		CustomerId:      uuid.New().String(),
		MerchantId:      uuid.New().String(),
		Amount:          &commongrpc.Money{Amount: "1000.00", Currency: "XAF"},
		PaymentType:     commongrpc.PaymentType_PAYMENT_TYPE_MMO,
		PayerPhoneNumber: "+237600000000",
		Provider:        commongrpc.Provider_PROVIDER_MTN_MOMO,
	})
	if err != nil {
		t.Fatalf("InitiateDeposit failed: %v", err)
	}
	if resp.Deposit == nil {
		t.Fatal("deposit should not be nil")
	}
}

func TestGetDeposit(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	depositRepo := mocks.NewMockDepositRepo(ctrl)
	customerRepo := mocks.NewMockCustomerRepo(ctrl)
	service := NewDepositService(depositRepo, customerRepo, zerolog.Nop())

	depositID := uuid.New()
	depositRepo.EXPECT().GetByID(gomock.Any(), depositID).
		Return(sqlc.Deposit{ID: depositID}, nil)

	resp, err := service.GetDeposit(context.Background(), &transactionsgrpc.GetDepositRequest{
		DepositId: depositID.String(),
	})
	if err != nil {
		t.Fatalf("GetDeposit failed: %v", err)
	}
	if resp.Deposit.Id != depositID.String() {
		t.Fatalf("deposit id = %s, want %s", resp.Deposit.Id, depositID.String())
	}
}

func TestGetDepositNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	depositRepo := mocks.NewMockDepositRepo(ctrl)
	customerRepo := mocks.NewMockCustomerRepo(ctrl)
	service := NewDepositService(depositRepo, customerRepo, zerolog.Nop())

	depositRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).
		Return(sqlc.Deposit{}, repo.ErrNotFound)

	_, err := service.GetDeposit(context.Background(), &transactionsgrpc.GetDepositRequest{
		DepositId: uuid.New().String(),
	})
	if got := status.Code(err); got != codes.NotFound {
		t.Fatalf("status code = %s, want %s", got, codes.NotFound)
	}
}

func TestGetDepositRepositoryError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	depositRepo := mocks.NewMockDepositRepo(ctrl)
	customerRepo := mocks.NewMockCustomerRepo(ctrl)
	service := NewDepositService(depositRepo, customerRepo, zerolog.Nop())

	depositRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).
		Return(sqlc.Deposit{}, errors.New("database down"))

	_, err := service.GetDeposit(context.Background(), &transactionsgrpc.GetDepositRequest{
		DepositId: uuid.New().String(),
	})
	if got := status.Code(err); got != codes.Internal {
		t.Fatalf("status code = %s, want %s", got, codes.Internal)
	}
}