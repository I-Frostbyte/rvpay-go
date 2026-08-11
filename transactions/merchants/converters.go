package merchants

import (
	transactionsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/sqlc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// merchantToProto maps a persisted merchant to its protobuf representation.
func merchantToProto(merchant sqlc.Merchant) *transactionsgrpc.Merchant {
	return &transactionsgrpc.Merchant{
		Id:        merchant.ID.String(),
		Name:      merchant.Name,
		Slug:      merchant.Slug,
		Status:    sqlcMerchantStatusToGrpc(merchant.Status),
		CreatedAt: timestamppb.New(merchant.CreatedAt),
		UpdatedAt: timestamppb.New(merchant.UpdatedAt),
	}
}

// sqlcMerchantStatusToGrpc maps a persisted merchant status to its protobuf
// representation. Unknown statuses map to the unspecified zero value.
func sqlcMerchantStatusToGrpc(merchantStatus sqlc.MerchantStatus) transactionsgrpc.MerchantStatus {
	switch merchantStatus {
	case sqlc.MerchantStatusONBOARDED:
		return transactionsgrpc.MerchantStatus_MERCHANT_STATUS_ONBOARDED
	case sqlc.MerchantStatusACTIVE:
		return transactionsgrpc.MerchantStatus_MERCHANT_STATUS_ACTIVE
	case sqlc.MerchantStatusSUSPENDED:
		return transactionsgrpc.MerchantStatus_MERCHANT_STATUS_SUSPENDED
	case sqlc.MerchantStatusRETIRED:
		return transactionsgrpc.MerchantStatus_MERCHANT_STATUS_RETIRED
	default:
		return transactionsgrpc.MerchantStatus_MERCHANT_STATUS_UNSPECIFIED
	}
}
