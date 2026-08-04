package deposits

import (
	"context"
	"fmt"

	"github.com/I-Frostbyte/pawapay_client"
	model "github.com/I-Frostbyte/pawapay_client/config"
	"github.com/I-Frostbyte/rvpay-go/deposits/db/repo"
	"github.com/I-Frostbyte/rvpay-go/deposits/db/sqlc"
	depositsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/depositsgrpc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Impl struct {
	repo          repo.DepositsRepo
	logger        zerolog.Logger
	pawapayClient pawapay_client.Client

	depositsgrpc.UnimplementedDepositsServiceServer
}

func NewDepositsService(
	repo repo.DepositsRepo,
	logger zerolog.Logger,
	pawapayClient pawapay_client.Client,
) *Impl {
	return &Impl{
		repo:          repo,
		logger:        logger,
		pawapayClient: pawapayClient,
	}
}

func (d *Impl) InitiateDeposit(ctx context.Context, req *depositsgrpc.CreateDepositRequest) (*depositsgrpc.CreateDepositResponse, error) {
	pawapay := d.pawapayClient

	var dbAmount pgtype.Numeric

	// Scan parses the string string representation (e.g., "1500.50") safely into the struct
	err := dbAmount.Scan(req.GetAmount())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse amount string to numeric: %v", err)
	}

	dummyClient, err := d.repo.Do().CreateClient(ctx, sqlc.CreateClientParams{
		ClientName:  "Socadel",
		Email:       "socadel+1@email.com",
		PhoneNumber: "+237699541235",
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not verify client for deposit: %v", err)
	}

	newDeposit, err := d.repo.Do().CreateDeposit(ctx, sqlc.CreateDepositParams{
		Amount: dbAmount,
		Currency: req.GetCurrency(),
		PayerType: grpcPayerTypeToSqlc(req.GetPayer().GetType()),
		PayerPhoneNumber: req.GetPayer().GetAccountDetails().GetPhoneNumber(),
		PayerProvider: grpcProviderToSqlc(req.GetPayer().GetAccountDetails().GetProvider()),
		ClientID: dummyClient.ID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not create deposit: %v", err)
	}

	payerType, err := sqlcPayerTypeToStringConverter(newDeposit.PayerType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not parse sqlc type to string (payer type): %v", err)
	}

	paymentProvider, err := sqlcPaymentProviderToStringConverter(newDeposit.PayerProvider)

	pawapayDeposit := model.Deposit{
		DepositID: newDeposit.ID.String(),
		Amount: float64(newDeposit.Amount.Exp),
		Currency: newDeposit.Currency,
		Payer: model.Payer{
			Type: payerType,
			AccountDetails: model.AccountDetails{
				PhoneNumber: newDeposit.PayerPhoneNumber,
				Provider: paymentProvider,
			},
		},
	}

	_, err = pawapay.DepositsService.InititateDeposit(&pawapayDeposit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not initiate deposit with pawapay client: %v", err)
	}

	return &depositsgrpc.CreateDepositResponse{
		DepositId: newDeposit.ID.String(),
		Status: depositsgrpc.DepositStatus_DEPOSIT_STATUS_ACCEPTED,
		NextStep: "FINAL_STATUS",
	}, nil
}

func sqlcPayerTypeToStringConverter(payerType sqlc.PayerType) (string, error) {
	var payerTypeString string
	switch payerType{
	case sqlc.PayerTypeMMO:
		payerTypeString = "MMO"
	default:
		return "", fmt.Errorf("Could not convert PayerType")
	}

	return payerTypeString, nil
}

func sqlcPaymentProviderToStringConverter(paymentProvider sqlc.PaymentProvider) (string, error) {
	var paymentProviderString string
	switch paymentProvider{
	case sqlc.PaymentProviderMTNMOMOCMR:
		paymentProviderString = "MTN_MOMO_CMR"
	case sqlc.PaymentProviderORANGECMR:
		paymentProviderString = "ORANGE_MOMO_CMR"
	default:
		return "", fmt.Errorf("Could not convert Payment Provider")
	}

	return paymentProviderString, nil
}

func grpcPayerTypeToSqlc(payerType depositsgrpc.DepositType) sqlc.PayerType {
	var sqlcType sqlc.PayerType

	switch payerType{
	case depositsgrpc.DepositType_DEPOSIT_PORTAL_MMO:
		sqlcType = sqlc.PayerTypeMMO
	default:
		sqlcType = sqlc.PayerTypeMMO
	}

	return sqlcType
}

func grpcProviderToSqlc(provider depositsgrpc.DepositProvider) sqlc.PaymentProvider {
	var sqlcProvider sqlc.PaymentProvider

	switch provider{
	case depositsgrpc.DepositProvider_DEPOSIT_PROVIDER_ORANGE_MOMO_CMR:
		sqlcProvider = sqlc.PaymentProviderORANGECMR
	default:
		sqlcProvider = sqlc.PaymentProviderMTNMOMOCMR
	}

	return sqlcProvider
}
