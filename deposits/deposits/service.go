package deposits

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/I-Frostbyte/pawapay_client"
	model "github.com/I-Frostbyte/pawapay_client/config"
	"github.com/I-Frostbyte/rvpay-go/deposits/db/repo"
	"github.com/I-Frostbyte/rvpay-go/deposits/db/sqlc"
	depositsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/depositsgrpc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "deposit request is required")
	}

	var dbAmount pgtype.Numeric

	// Scan parses the string representation (for example, "1500.50") safely into the struct.
	err := dbAmount.Scan(req.GetAmount())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid deposit amount: %v", err)
	}

	amount, err := dbAmount.Float64Value()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid deposit amount: %v", err)
	}
	if !amount.Valid || amount.Float64 <= 0 {
		return nil, status.Error(codes.InvalidArgument, "deposit amount must be greater than zero")
	}

	clientID, err := uuid.Parse(req.GetClientId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "client_id must be a valid UUID")
	}

	payer := req.GetPayer()
	if payer == nil || payer.GetAccountDetails() == nil {
		return nil, status.Error(codes.InvalidArgument, "payer and payer account details are required")
	}
	if strings.TrimSpace(payer.GetAccountDetails().GetPhoneNumber()) == "" {
		return nil, status.Error(codes.InvalidArgument, "payer phone number is required")
	}

	dbPayerType, err := grpcPayerTypeToSqlc(payer.GetType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	payerProvider, err := grpcProviderToSqlc(payer.GetAccountDetails().GetProvider())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	queries := d.repo.Do()
	client, err := queries.GetClientByID(ctx, clientID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "client not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not verify client for deposit: %v", err)
	}

	newDeposit, err := queries.CreateDeposit(ctx, sqlc.CreateDepositParams{
		Amount:           dbAmount,
		Currency:         req.GetCurrency(),
		PayerType:        dbPayerType,
		PayerPhoneNumber: payer.GetAccountDetails().GetPhoneNumber(),
		PayerProvider:    payerProvider,
		ClientID:         client.ID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not create deposit: %v", err)
	}

	payerType, err := sqlcPayerTypeToStringConverter(newDeposit.PayerType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not parse sqlc type to string (payer type): %v", err)
	}

	paymentProvider, err := sqlcPaymentProviderToStringConverter(newDeposit.PayerProvider)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not parse sqlc type to string (payment provider): %v", err)
	}

	pawapayDeposit := model.Deposit{
		DepositID: newDeposit.ID.String(),
		Amount:    amount.Float64,
		Currency:  newDeposit.Currency,
		Payer: model.Payer{
			Type: payerType,
			AccountDetails: model.AccountDetails{
				PhoneNumber: newDeposit.PayerPhoneNumber,
				Provider:    paymentProvider,
			},
		},
	}

	_, err = pawapay.DepositsService.InititateDeposit(&pawapayDeposit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not initiate deposit with pawapay client: %v", err)
	}

	return &depositsgrpc.CreateDepositResponse{
		DepositId: newDeposit.ID.String(),
		Status:    depositsgrpc.DepositStatus_DEPOSIT_STATUS_ACCEPTED,
		NextStep:  "FINAL_STATUS",
	}, nil
}

func sqlcPayerTypeToStringConverter(payerType sqlc.PayerType) (string, error) {
	var payerTypeString string
	switch payerType {
	case sqlc.PayerTypeMMO:
		payerTypeString = "MMO"
	default:
		return "", fmt.Errorf("Could not convert PayerType")
	}

	return payerTypeString, nil
}

func sqlcPaymentProviderToStringConverter(paymentProvider sqlc.PaymentProvider) (string, error) {
	var paymentProviderString string
	switch paymentProvider {
	case sqlc.PaymentProviderMTNMOMOCMR:
		paymentProviderString = "MTN_MOMO_CMR"
	case sqlc.PaymentProviderORANGECMR:
		paymentProviderString = "ORANGE_MOMO_CMR"
	default:
		return "", fmt.Errorf("Could not convert Payment Provider")
	}

	return paymentProviderString, nil
}

func grpcPayerTypeToSqlc(payerType depositsgrpc.DepositType) (sqlc.PayerType, error) {
	switch payerType {
	case depositsgrpc.DepositType_DEPOSIT_PORTAL_MMO:
		return sqlc.PayerTypeMMO, nil
	default:
		return "", fmt.Errorf("unsupported payer type: %s", payerType)
	}
}

func grpcProviderToSqlc(provider depositsgrpc.DepositProvider) (sqlc.PaymentProvider, error) {
	switch provider {
	case depositsgrpc.DepositProvider_DEPOSIT_PROVIDER_MTN_MOMO_CMR:
		return sqlc.PaymentProviderMTNMOMOCMR, nil
	case depositsgrpc.DepositProvider_DEPOSIT_PROVIDER_ORANGE_MOMO_CMR:
		return sqlc.PaymentProviderORANGECMR, nil
	default:
		return "", fmt.Errorf("unsupported payer provider: %s", provider)
	}
}
