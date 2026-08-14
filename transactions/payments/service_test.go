package payments

import (
	"context"
	"errors"
	"testing"

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

func TestVerifyPaymentValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  *transactionsgrpc.VerifyPaymentRequest
		code codes.Code
	}{
		{name: "missing request", code: codes.InvalidArgument},
		{name: "missing transaction id", req: &transactionsgrpc.VerifyPaymentRequest{}, code: codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			depositRepo := mocks.NewMockDepositRepo(ctrl)
			service := NewPaymentService(depositRepo, zerolog.Nop())

			_, err := service.VerifyPayment(context.Background(), tt.req)
			if got := status.Code(err); got != tt.code {
				t.Fatalf("status code = %s, want %s", got, tt.code)
			}
		})
	}
}

func TestVerifyPaymentCompleted(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	depositRepo := mocks.NewMockDepositRepo(ctrl)
	service := NewPaymentService(depositRepo, zerolog.Nop())

	depositRepo.EXPECT().GetByGHLTransactionID(gomock.Any(), "txn-1").
		Return(sqlc.Deposit{ID: uuid.New(), Status: sqlc.DepositStatusCOMPLETED}, nil)

	resp, err := service.VerifyPayment(context.Background(), &transactionsgrpc.VerifyPaymentRequest{
		GhlTransactionId: "txn-1",
	})
	if err != nil {
		t.Fatalf("VerifyPayment failed: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected success=true for completed deposit")
	}
	if resp.Failed {
		t.Fatal("expected failed=false for completed deposit")
	}
}

func TestVerifyPaymentFailed(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	depositRepo := mocks.NewMockDepositRepo(ctrl)
	service := NewPaymentService(depositRepo, zerolog.Nop())

	depositRepo.EXPECT().GetByGHLTransactionID(gomock.Any(), "txn-1").
		Return(sqlc.Deposit{ID: uuid.New(), Status: sqlc.DepositStatusFAILED}, nil)

	resp, err := service.VerifyPayment(context.Background(), &transactionsgrpc.VerifyPaymentRequest{
		GhlTransactionId: "txn-1",
	})
	if err != nil {
		t.Fatalf("VerifyPayment failed: %v", err)
	}
	if !resp.Failed {
		t.Fatal("expected failed=true for failed deposit")
	}
	if resp.Success {
		t.Fatal("expected success=false for failed deposit")
	}
}

func TestVerifyPaymentPending(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	depositRepo := mocks.NewMockDepositRepo(ctrl)
	service := NewPaymentService(depositRepo, zerolog.Nop())

	depositRepo.EXPECT().GetByGHLTransactionID(gomock.Any(), "txn-1").
		Return(sqlc.Deposit{ID: uuid.New(), Status: sqlc.DepositStatusINITIATED}, nil)

	resp, err := service.VerifyPayment(context.Background(), &transactionsgrpc.VerifyPaymentRequest{
		GhlTransactionId: "txn-1",
	})
	if err != nil {
		t.Fatalf("VerifyPayment failed: %v", err)
	}
	if resp.Success {
		t.Fatal("expected success=false for pending deposit")
	}
	if resp.Failed {
		t.Fatal("expected failed=false for pending deposit")
	}
}

func TestVerifyPaymentNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	depositRepo := mocks.NewMockDepositRepo(ctrl)
	service := NewPaymentService(depositRepo, zerolog.Nop())

	depositRepo.EXPECT().GetByGHLTransactionID(gomock.Any(), "unknown-txn").
		Return(sqlc.Deposit{}, repo.ErrNotFound)

	_, err := service.VerifyPayment(context.Background(), &transactionsgrpc.VerifyPaymentRequest{
		GhlTransactionId: "unknown-txn",
	})
	if got := status.Code(err); got != codes.NotFound {
		t.Fatalf("status code = %s, want %s", got, codes.NotFound)
	}
}

func TestVerifyPaymentRepositoryError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	depositRepo := mocks.NewMockDepositRepo(ctrl)
	service := NewPaymentService(depositRepo, zerolog.Nop())

	depositRepo.EXPECT().GetByGHLTransactionID(gomock.Any(), "txn-1").
		Return(sqlc.Deposit{}, errors.New("database down"))

	_, err := service.VerifyPayment(context.Background(), &transactionsgrpc.VerifyPaymentRequest{
		GhlTransactionId: "txn-1",
	})
	if got := status.Code(err); got != codes.Internal {
		t.Fatalf("status code = %s, want %s", got, codes.Internal)
	}
}

func TestProcessPaymentWebhookValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  *transactionsgrpc.ProcessPaymentWebhookRequest
		code codes.Code
	}{
		{name: "missing request", code: codes.InvalidArgument},
		{name: "missing event type", req: &transactionsgrpc.ProcessPaymentWebhookRequest{}, code: codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			depositRepo := mocks.NewMockDepositRepo(ctrl)
			service := NewPaymentService(depositRepo, zerolog.Nop())

			_, err := service.ProcessPaymentWebhook(context.Background(), tt.req)
			if got := status.Code(err); got != tt.code {
				t.Fatalf("status code = %s, want %s", got, tt.code)
			}
		})
	}
}

func TestProcessPaymentWebhookUnknownEventType(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	depositRepo := mocks.NewMockDepositRepo(ctrl)
	service := NewPaymentService(depositRepo, zerolog.Nop())

	// Unknown event types are acknowledged safely without processing.
	_, err := service.ProcessPaymentWebhook(context.Background(), &transactionsgrpc.ProcessPaymentWebhookRequest{
		EventType: "subscription.active",
	})
	if err != nil {
		t.Fatalf("unexpected error for unknown event type: %v", err)
	}
}

func TestProcessPaymentWebhookMissingTransactionID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	depositRepo := mocks.NewMockDepositRepo(ctrl)
	service := NewPaymentService(depositRepo, zerolog.Nop())

	// payment.captured without a transaction_id is acknowledged safely.
	_, err := service.ProcessPaymentWebhook(context.Background(), &transactionsgrpc.ProcessPaymentWebhookRequest{
		EventType: "payment.captured",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessPaymentWebhookPaymentCaptured(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	depositRepo := mocks.NewMockDepositRepo(ctrl)
	service := NewPaymentService(depositRepo, zerolog.Nop())

	depositID := uuid.New()
	depositRepo.EXPECT().GetByGHLTransactionID(gomock.Any(), "txn-1").
		Return(sqlc.Deposit{ID: depositID, Status: sqlc.DepositStatusCOMPLETED}, nil)
	depositRepo.EXPECT().UpdateGHLReference(gomock.Any(), depositID, "txn-1", "charge-1").
		Return(sqlc.Deposit{ID: depositID}, nil)

	_, err := service.ProcessPaymentWebhook(context.Background(), &transactionsgrpc.ProcessPaymentWebhookRequest{
		EventType:     "payment.captured",
		TransactionId: "txn-1",
		ChargeId:      "charge-1",
		LocationId:    "loc-1",
	})
	if err != nil {
		t.Fatalf("ProcessPaymentWebhook failed: %v", err)
	}
}

func TestProcessPaymentWebhookUnknownTransaction(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	depositRepo := mocks.NewMockDepositRepo(ctrl)
	service := NewPaymentService(depositRepo, zerolog.Nop())

	depositRepo.EXPECT().GetByGHLTransactionID(gomock.Any(), "unknown-txn").
		Return(sqlc.Deposit{}, repo.ErrNotFound)

	// Unknown transaction is acknowledged safely; the event is already
	// recorded for idempotency by the transport adapter.
	_, err := service.ProcessPaymentWebhook(context.Background(), &transactionsgrpc.ProcessPaymentWebhookRequest{
		EventType:     "payment.captured",
		TransactionId: "unknown-txn",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
