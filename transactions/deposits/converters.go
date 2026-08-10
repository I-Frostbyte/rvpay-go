package deposits

import (
	"strconv"

	"github.com/I-Frostbyte/rvpay-go/transactions/db/sqlc"
	commongrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/commongrpc"
	transactionsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// depositToProto maps a persisted deposit to its protobuf representation.
func depositToProto(deposit sqlc.Deposit) *transactionsgrpc.Deposit {
	proto := &transactionsgrpc.Deposit{
		Id:               deposit.ID.String(),
		ClientId:         deposit.ClientID.String(),
		CustomerId:       deposit.CustomerID.String(),
		MerchantId:       deposit.MerchantID.String(),
		Amount:           &commongrpc.Money{},
		PaymentType:      sqlcPaymentTypeToGrpc(deposit.PaymentType),
		PayerPhoneNumber: deposit.PayerPhoneNumber,
		Provider:         sqlcPaymentProviderToGrpc(deposit.Provider),
		Status:           sqlcDepositStatusToGrpc(deposit.Status),
		ExternalReference: deposit.ExternalReference,
		InitiatedAt:      timestamppb.New(deposit.InitiatedAt),
		CreatedAt:        timestamppb.New(deposit.CreatedAt),
		UpdatedAt:        timestamppb.New(deposit.UpdatedAt),
	}

	// Amount is stored as NUMERIC(18,2); expose it as a decimal string.
	if amount, err := deposit.Amount.Float64Value(); err == nil && amount.Valid {
		proto.Amount.Amount = strconv.FormatFloat(amount.Float64, 'f', 2, 64)
	}
	proto.Amount.Currency = deposit.Currency

	// Optional lifecycle timestamps.
	if deposit.CompletedAt.Valid {
		proto.CompletedAt = timestamppb.New(deposit.CompletedAt.Time)
	}
	if deposit.FailedAt.Valid {
		proto.FailedAt = timestamppb.New(deposit.FailedAt.Time)
	}
	proto.FailureReason = deposit.FailureReason

	return proto
}

// sqlcDepositStatusToGrpc maps a persisted deposit status to its protobuf
// representation. Unknown statuses map to the unspecified zero value.
func sqlcDepositStatusToGrpc(depositStatus sqlc.DepositStatus) transactionsgrpc.DepositStatus {
	switch depositStatus {
	case sqlc.DepositStatusINITIATED:
		return transactionsgrpc.DepositStatus_DEPOSIT_STATUS_INITIATED
	case sqlc.DepositStatusPROCESSING:
		return transactionsgrpc.DepositStatus_DEPOSIT_STATUS_PROCESSING
	case sqlc.DepositStatusCOMPLETED:
		return transactionsgrpc.DepositStatus_DEPOSIT_STATUS_COMPLETED
	case sqlc.DepositStatusFAILED:
		return transactionsgrpc.DepositStatus_DEPOSIT_STATUS_FAILED
	default:
		return transactionsgrpc.DepositStatus_DEPOSIT_STATUS_UNSPECIFIED
	}
}

// sqlcPaymentTypeToGrpc maps a persisted payment type to its protobuf
// representation. Unknown types map to the unspecified zero value.
func sqlcPaymentTypeToGrpc(paymentType sqlc.PaymentType) commongrpc.PaymentType {
	switch paymentType {
	case sqlc.PaymentTypeMMO:
		return commongrpc.PaymentType_PAYMENT_TYPE_MMO
	case sqlc.PaymentTypeCREDITCARD:
		return commongrpc.PaymentType_PAYMENT_TYPE_CREDIT_CARD
	default:
		return commongrpc.PaymentType_PAYMENT_TYPE_UNSPECIFIED
	}
}

// sqlcPaymentProviderToGrpc maps a persisted payment provider to its protobuf
// representation. Unknown providers map to the unspecified zero value.
func sqlcPaymentProviderToGrpc(paymentProvider sqlc.PaymentProvider) commongrpc.Provider {
	switch paymentProvider {
	case sqlc.PaymentProviderMTNMOMO:
		return commongrpc.Provider_PROVIDER_MTN_MOMO
	case sqlc.PaymentProviderORANGEMOMO:
		return commongrpc.Provider_PROVIDER_ORANGE_MOMO
	default:
		return commongrpc.Provider_PROVIDER_UNSPECIFIED
	}
}