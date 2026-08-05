package deposits

import (
	"context"
	"testing"

	"github.com/I-Frostbyte/pawapay_client"
	depositsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/depositsgrpc"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestInitiateDepositRejectsInvalidRequests(t *testing.T) {
	t.Parallel()

	service := NewDepositsService(nil, zerolog.Nop(), pawapay_client.Client{})
	validPayer := &depositsgrpc.Participant{
		Type: depositsgrpc.DepositType_DEPOSIT_PORTAL_MMO,
		AccountDetails: &depositsgrpc.AccountDetails{
			PhoneNumber: "+237699541235",
			Provider:    depositsgrpc.DepositProvider_DEPOSIT_PROVIDER_MTN_MOMO_CMR,
		},
	}

	tests := []struct {
		name string
		req  *depositsgrpc.CreateDepositRequest
		code codes.Code
	}{
		{name: "missing request", code: codes.InvalidArgument},
		{name: "invalid amount", req: &depositsgrpc.CreateDepositRequest{Amount: "invalid"}, code: codes.InvalidArgument},
		{name: "zero amount", req: &depositsgrpc.CreateDepositRequest{Amount: "0"}, code: codes.InvalidArgument},
		{name: "invalid client ID", req: &depositsgrpc.CreateDepositRequest{Amount: "1", ClientId: "invalid"}, code: codes.InvalidArgument},
		{name: "missing payer", req: &depositsgrpc.CreateDepositRequest{Amount: "1", ClientId: "0e8caa3c-77fb-4e69-9241-79a8a9be5bdb"}, code: codes.InvalidArgument},
		{name: "missing phone number", req: &depositsgrpc.CreateDepositRequest{Amount: "1", ClientId: "0e8caa3c-77fb-4e69-9241-79a8a9be5bdb", Payer: &depositsgrpc.Participant{AccountDetails: &depositsgrpc.AccountDetails{}}}, code: codes.InvalidArgument},
		{name: "unsupported payer type", req: &depositsgrpc.CreateDepositRequest{Amount: "1", ClientId: "0e8caa3c-77fb-4e69-9241-79a8a9be5bdb", Payer: &depositsgrpc.Participant{Type: depositsgrpc.DepositType_DEPOSIT_PORTAL_CARD, AccountDetails: validPayer.GetAccountDetails()}}, code: codes.InvalidArgument},
		{name: "unsupported provider", req: &depositsgrpc.CreateDepositRequest{Amount: "1", ClientId: "0e8caa3c-77fb-4e69-9241-79a8a9be5bdb", Payer: &depositsgrpc.Participant{Type: validPayer.GetType(), AccountDetails: &depositsgrpc.AccountDetails{PhoneNumber: validPayer.GetAccountDetails().GetPhoneNumber()}}}, code: codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := service.InitiateDeposit(context.Background(), tt.req)
			if got := status.Code(err); got != tt.code {
				t.Fatalf("status code = %s, want %s", got, tt.code)
			}
		})
	}
}

func TestGrpcPayerTypeToSqlc(t *testing.T) {
	t.Parallel()

	if _, err := grpcPayerTypeToSqlc(depositsgrpc.DepositType_DEPOSIT_PORTAL_MMO); err != nil {
		t.Fatalf("MMO payer type returned error: %v", err)
	}
	if _, err := grpcPayerTypeToSqlc(depositsgrpc.DepositType_DEPOSIT_PORTAL_CARD); err == nil {
		t.Fatal("CARD payer type did not return an error")
	}
}

func TestGrpcProviderToSqlc(t *testing.T) {
	t.Parallel()

	for _, provider := range []depositsgrpc.DepositProvider{
		depositsgrpc.DepositProvider_DEPOSIT_PROVIDER_MTN_MOMO_CMR,
		depositsgrpc.DepositProvider_DEPOSIT_PROVIDER_ORANGE_MOMO_CMR,
	} {
		if _, err := grpcProviderToSqlc(provider); err != nil {
			t.Fatalf("provider %s returned error: %v", provider, err)
		}
	}
	if _, err := grpcProviderToSqlc(depositsgrpc.DepositProvider_DEPOSIT_PROVIDER_UNSPECIFIED); err == nil {
		t.Fatal("unspecified provider did not return an error")
	}
}
