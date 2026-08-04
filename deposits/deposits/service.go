package deposits

import (
	"context"

	"github.com/I-Frostbyte/rvpay-go/deposits/db/repo"
	depositsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/depositsgrpc"
	"github.com/rs/zerolog"
)

type Impl struct {
	repo   repo.DepositsRepo
	logger zerolog.Logger

	depositsgrpc.UnimplementedDepositsServiceServer
}

func NewDepositsService(
	repo repo.DepositsRepo,
	logger zerolog.Logger,
) *Impl {
	return &Impl{
		repo:   repo,
		logger: logger,
	}
}

func (d *Impl) InitiateDeposit(ctx context.Context, req *depositsgrpc.CreateDepositRequest) (*depositsgrpc.CreateDepositResponse, error) {
	panic("unimplemented")
}
