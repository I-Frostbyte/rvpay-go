package customers

import (
	transactionsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc"
	"github.com/I-Frostbyte/rvpay-go/transactions/db/sqlc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// customerToProto maps a persisted customer to its protobuf representation.
func customerToProto(customer sqlc.Customer) *transactionsgrpc.Customer {
	return &transactionsgrpc.Customer{
		Id:          customer.ID.String(),
		ClientId:    customer.ClientID.String(),
		MerchantId:  customer.MerchantID.String(),
		PhoneNumber: customer.PhoneNumber,
		Status:      sqlcCustomerStatusToGrpc(customer.Status),
		CreatedAt:   timestamppb.New(customer.CreatedAt),
		UpdatedAt:   timestamppb.New(customer.UpdatedAt),
	}
}

// sqlcCustomerStatusToGrpc maps a persisted customer status to its protobuf
// representation. Unknown statuses map to the unspecified zero value.
func sqlcCustomerStatusToGrpc(customerStatus sqlc.CustomerStatus) transactionsgrpc.CustomerStatus {
	switch customerStatus {
	case sqlc.CustomerStatusCREATED:
		return transactionsgrpc.CustomerStatus_CUSTOMER_STATUS_CREATED
	case sqlc.CustomerStatusACTIVE:
		return transactionsgrpc.CustomerStatus_CUSTOMER_STATUS_ACTIVE
	default:
		return transactionsgrpc.CustomerStatus_CUSTOMER_STATUS_UNSPECIFIED
	}
}
