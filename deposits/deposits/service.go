package deposits

import (
	"context"

	"github.com/rs/zerolog"
	depositsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/depositsgrpc"
)

type Impl struct {
	logger zerolog.Logger

	depositsgrpc.UnimplementedDepositsServiceServer
}

func NewDepositsService(
	logger zerolog.Logger,
) *Impl {
	return &Impl{
		logger: logger,
	}
}


func (d *Impl) InitiateDeposit(ctx context.Context, req *depositsgrpc.CreateDepositRequest) (*depositsgrpc.CreateDepositResponse, error) {
	panic("unimplemented")
}