package payouts

import (
	"strconv"

	"github.com/I-Frostbyte/rvpay-go/transactions/db/sqlc"
	commongrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/commongrpc"
	transactionsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// payoutToProto maps a persisted payout to its protobuf representation.
func payoutToProto(payout sqlc.Payout) *transactionsgrpc.Payout {
	proto := &transactionsgrpc.Payout{
		Id:                  payout.ID.String(),
		ClientId:            payout.ClientID.String(),
		MerchantId:          payout.MerchantID.String(),
		Amount:              &commongrpc.Money{},
		Provider:            sqlcPaymentProviderToGrpc(payout.Provider),
		DestinationReference: payout.DestinationReference,
		Status:              sqlcPayoutStatusToGrpc(payout.Status),
		ExternalReference:   payout.ExternalReference,
		RequestedAt:         timestamppb.New(payout.RequestedAt),
		CreatedAt:           timestamppb.New(payout.CreatedAt),
		UpdatedAt:           timestamppb.New(payout.UpdatedAt),
	}

	// Amount is stored as NUMERIC(18,2); expose it as a decimal string.
	if amount, err := payout.Amount.Float64Value(); err == nil && amount.Valid {
		proto.Amount.Amount = strconv.FormatFloat(amount.Float64, 'f', 2, 64)
	}
	proto.Amount.Currency = payout.Currency

	// Optional lifecycle timestamps.
	if payout.CompletedAt.Valid {
		proto.CompletedAt = timestamppb.New(payout.CompletedAt.Time)
	}
	if payout.FailedAt.Valid {
		proto.FailedAt = timestamppb.New(payout.FailedAt.Time)
	}
	proto.FailureReason = payout.FailureReason

	return proto
}

// sqlcPayoutStatusToGrpc maps a persisted payout status to its protobuf
// representation. Unknown statuses map to the unspecified zero value.
func sqlcPayoutStatusToGrpc(payoutStatus sqlc.PayoutStatus) transactionsgrpc.PayoutStatus {
	switch payoutStatus {
	case sqlc.PayoutStatusREQUESTED:
		return transactionsgrpc.PayoutStatus_PAYOUT_STATUS_REQUESTED
	case sqlc.PayoutStatusPROCESSING:
		return transactionsgrpc.PayoutStatus_PAYOUT_STATUS_PROCESSING
	case sqlc.PayoutStatusCOMPLETED:
		return transactionsgrpc.PayoutStatus_PAYOUT_STATUS_COMPLETED
	case sqlc.PayoutStatusFAILED:
		return transactionsgrpc.PayoutStatus_PAYOUT_STATUS_FAILED
	default:
		return transactionsgrpc.PayoutStatus_PAYOUT_STATUS_UNSPECIFIED
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