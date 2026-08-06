package integration

import (
	"fmt"

	"github.com/I-Frostbyte/rvpay-go/integrations/db/sqlc"
	integrationsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/integrationsgrpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func sqlcIntegrationToGrpc(integration sqlc.Integration) *integrationsgrpc.Integration {
	return &integrationsgrpc.Integration{
		Id:             integration.ID.String(),
		Provider:       integration.Provider,
		LocationId:     integration.LocationID,
		AccessToken:    integration.AccessToken,
		RefreshToken:   integration.RefreshToken,
		TokenExpiresAt: timestamppb.New(integration.TokenExpiresAt),
		CreatedAt:      timestamppb.New(integration.CreatedAt),
		UpdatedAt:      timestamppb.New(integration.UpdatedAt),
	}
}

func grpcIntegrationRequestToSqlc(req *integrationsgrpc.CreateIntegrationRequest) (sqlc.CreateIntegrationParams, error) {
	if req == nil {
		return sqlc.CreateIntegrationParams{}, fmt.Errorf("integration request is required")
	}

	if req.GetTokenExpiresAt() == nil {
		return sqlc.CreateIntegrationParams{}, fmt.Errorf("token expires at is required")
	}

	return sqlc.CreateIntegrationParams{
		Provider:       req.GetProvider(),
		LocationID:     req.GetLocationId(),
		AccessToken:    req.GetAccessToken(),
		RefreshToken:   req.GetRefreshToken(),
		TokenExpiresAt: req.GetTokenExpiresAt().AsTime(),
	}, nil
}