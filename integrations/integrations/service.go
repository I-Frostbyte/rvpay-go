package integrations

import (
	"context"
	"errors"

	integrationsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/integrationsgrpc"
	"github.com/I-Frostbyte/rvpay-go/integrations/db/repo"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Impl struct {
	repo   repo.IntegrationsRepo
	logger zerolog.Logger

	integrationsgrpc.UnimplementedIntegrationServiceServer
}

func NewIntegrationService(
	repo repo.IntegrationsRepo,
	logger zerolog.Logger,
) *Impl {
	return &Impl{
		repo:   repo,
		logger: logger,
	}
}

func (i *Impl) CreateIntegration(ctx context.Context, req *integrationsgrpc.CreateIntegrationRequest) (*integrationsgrpc.CreateIntegrationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "integration request is required")
	}

	createParams, err := grpcIntegrationRequestToSqlc(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	queries := i.repo.Do()
	integration, err := queries.CreateIntegration(ctx, createParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not create integration: %v", err)
	}

	return &integrationsgrpc.CreateIntegrationResponse{
		Integration: sqlcIntegrationToGrpc(integration),
	}, nil
}

func (i *Impl) GetIntegration(ctx context.Context, req *integrationsgrpc.GetIntegrationRequest) (*integrationsgrpc.GetIntegrationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "integration request is required")
	}

	integrationID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id must be a valid UUID")
	}

	queries := i.repo.Do()
	integration, err := queries.GetIntegrationByID(ctx, integrationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "integration not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get integration: %v", err)
	}

	return &integrationsgrpc.GetIntegrationResponse{
		Integration: sqlcIntegrationToGrpc(integration),
	}, nil
}

func (i *Impl) DeleteIntegration(ctx context.Context, req *integrationsgrpc.DeleteIntegrationRequest) (*integrationsgrpc.DeleteIntegrationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "integration request is required")
	}

	integrationID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id must be a valid UUID")
	}

	queries := i.repo.Do()
	rowsAffected, err := queries.DeleteIntegration(ctx, integrationID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not delete integration: %v", err)
	}
	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "integration not found")
	}

	return &integrationsgrpc.DeleteIntegrationResponse{
		Id: req.GetId(),
	}, nil
}
