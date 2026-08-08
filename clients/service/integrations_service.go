package service

import (
	"context"
	"errors"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	clientsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/clientsgrpc"
	"github.com/google/uuid"
)

type IntegrationsServiceImpl struct {
	integrationsRepo repo.IntegrationRepo
	clientsRepo      repo.ClientRepo
	platformsRepo    repo.PlatformRepo
	oauthRepo        repo.OAuthTokenRepo
	webhookRepo      repo.WebhookSubscriptionRepo
	logger           Logger
}

func NewIntegrationsServiceImpl(
	integrationsRepo repo.IntegrationRepo,
	clientsRepo repo.ClientRepo,
	platformsRepo repo.PlatformRepo,
	oauthRepo repo.OAuthTokenRepo,
	webhookRepo repo.WebhookSubscriptionRepo,
	logger Logger,
) *IntegrationsServiceImpl {
	return &IntegrationsServiceImpl{
		integrationsRepo: integrationsRepo,
		clientsRepo:      clientsRepo,
		platformsRepo:    platformsRepo,
		oauthRepo:        oauthRepo,
		webhookRepo:      webhookRepo,
		logger:           logger,
	}
}

func (s *IntegrationsServiceImpl) InstallIntegration(ctx context.Context, req *clientsgrpc.InstallIntegrationRequest) (*clientsgrpc.InstallIntegrationResponse, error) {
	clientID, err := parseUUID(req.GetClientId())
	if err != nil {
		return nil, err
	}
	platformID, err := parseUUID(req.GetPlatformId())
	if err != nil {
		return nil, err
	}

	client, err := s.clientsRepo.GetByID(ctx, clientID)
	if err == repo.ErrNotFound {
		return nil, ErrClientNotFound
	}
	if err != nil {
		return nil, translateRepoError(err)
	}

	if client.Status != sqlc.ClientStatusACTIVE {
		return nil, ErrClientInactive
	}

	platform, err := s.platformsRepo.GetByID(ctx, platformID)
	if err == repo.ErrNotFound {
		return nil, ErrPlatformNotFound
	}
	if err != nil {
		return nil, translateRepoError(err)
	}

	if !platform.Enabled {
		return nil, ErrPlatformDisabled
	}

	_, err = s.integrationsRepo.GetByClientAndPlatform(ctx, clientID, platformID)
	if err == nil {
		return nil, ErrIntegrationAlreadyExists
	}
	if !errors.Is(err, repo.ErrNotFound) {
		return nil, translateRepoError(err)
	}

	integration, err := s.integrationsRepo.Create(ctx, clientID, platformID, "", sqlc.IntegrationStatusCREATED)
	if err != nil {
		return nil, translateRepoError(err)
	}

	s.logger.Info("integration installed", "integration_id", integration.ID, "client_id", clientID, "platform_id", platformID)

	return &clientsgrpc.InstallIntegrationResponse{
		Integration: sqlcIntegrationToProto(integration),
	}, nil
}

func (s *IntegrationsServiceImpl) UninstallIntegration(ctx context.Context, req *clientsgrpc.UninstallIntegrationRequest) (*clientsgrpc.UninstallIntegrationResponse, error) {
	id, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	integration, err := s.integrationsRepo.GetByID(ctx, id)
	if err == repo.ErrNotFound {
		return nil, ErrIntegrationNotFound
	}
	if err != nil {
		return nil, translateRepoError(err)
	}

	if integration.Status == sqlc.IntegrationStatusACTIVE {
		_, err := s.integrationsRepo.UpdateStatus(ctx, id, sqlc.IntegrationStatusREVOKED)
		if err != nil {
			return nil, translateRepoError(err)
		}
	}

	err = s.integrationsRepo.Delete(ctx, id)
	if err != nil {
		return nil, translateRepoError(err)
	}

	s.logger.Info("integration uninstalled", "integration_id", id)

	return &clientsgrpc.UninstallIntegrationResponse{
		Id: id.String(),
	}, nil
}

func (s *IntegrationsServiceImpl) GetIntegration(ctx context.Context, req *clientsgrpc.GetIntegrationRequest) (*clientsgrpc.GetIntegrationResponse, error) {
	id, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	integration, err := s.integrationsRepo.GetByID(ctx, id)
	if err == repo.ErrNotFound {
		return nil, ErrIntegrationNotFound
	}
	if err != nil {
		return nil, translateRepoError(err)
	}

	return &clientsgrpc.GetIntegrationResponse{
		Integration: sqlcIntegrationToProto(integration),
	}, nil
}

func (s *IntegrationsServiceImpl) ListIntegrations(ctx context.Context, req *clientsgrpc.ListIntegrationsRequest) (*clientsgrpc.ListIntegrationsResponse, error) {
	pageSize := int32(20)
	offset := int32(0)
	if req.GetPagination() != nil {
		if req.GetPagination().PageSize > 0 {
			pageSize = req.GetPagination().PageSize
		}
		if req.GetPagination().PageToken != "" {
			offset = pageSize
		}
	}

	var integrations []sqlc.Integration
	var err error

	clientID, err := parseUUID(req.GetClientId())
	if err == nil {
		integrations, err = s.integrationsRepo.ListByClient(ctx, clientID, pageSize, offset)
	} else {
		integrations, err = s.integrationsRepo.ListByClient(ctx, uuid.Nil, pageSize, offset)
	}
	if err != nil {
		return nil, translateRepoError(err)
	}

	protoIntegrations := make([]*clientsgrpc.Integration, 0, len(integrations))
	for _, i := range integrations {
		protoIntegrations = append(protoIntegrations, sqlcIntegrationToProto(i))
	}

	return &clientsgrpc.ListIntegrationsResponse{
		Integrations: protoIntegrations,
	}, nil
}

func (s *IntegrationsServiceImpl) ReconnectIntegration(ctx context.Context, req *clientsgrpc.ReconnectIntegrationRequest) (*clientsgrpc.ReconnectIntegrationResponse, error) {
	id, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	integration, err := s.integrationsRepo.GetByID(ctx, id)
	if err == repo.ErrNotFound {
		return nil, ErrIntegrationNotFound
	}
	if err != nil {
		return nil, translateRepoError(err)
	}

	if integration.Status != sqlc.IntegrationStatusACTIVE {
		return nil, ErrIntegrationNotActive
	}

	updated, err := s.integrationsRepo.UpdateStatus(ctx, id, sqlc.IntegrationStatusCREATED)
	if err != nil {
		return nil, translateRepoError(err)
	}

	s.logger.Info("integration reconnected", "integration_id", updated.ID)

	return &clientsgrpc.ReconnectIntegrationResponse{
		Integration: sqlcIntegrationToProto(updated),
	}, nil
}

func (s *IntegrationsServiceImpl) DisconnectIntegration(ctx context.Context, req *clientsgrpc.DisconnectIntegrationRequest) (*clientsgrpc.DisconnectIntegrationResponse, error) {
	id, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	_, err = s.integrationsRepo.GetByID(ctx, id)
	if err == repo.ErrNotFound {
		return nil, ErrIntegrationNotFound
	}
	if err != nil {
		return nil, translateRepoError(err)
	}

	updated, err := s.integrationsRepo.UpdateStatus(ctx, id, sqlc.IntegrationStatusREVOKED)
	if err != nil {
		return nil, translateRepoError(err)
	}

	s.logger.Info("integration disconnected", "integration_id", updated.ID)

	return &clientsgrpc.DisconnectIntegrationResponse{
		Integration: sqlcIntegrationToProto(updated),
	}, nil
}

func (s *IntegrationsServiceImpl) SyncIntegration(ctx context.Context, req *clientsgrpc.SyncIntegrationRequest) (*clientsgrpc.SyncIntegrationResponse, error) {
	id, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	integration, err := s.integrationsRepo.GetByID(ctx, id)
	if err == repo.ErrNotFound {
		return nil, ErrIntegrationNotFound
	}
	if err != nil {
		return nil, translateRepoError(err)
	}

	if integration.Status != sqlc.IntegrationStatusACTIVE {
		return nil, ErrIntegrationNotActive
	}

	updated, err := s.integrationsRepo.UpdateStatus(ctx, id, sqlc.IntegrationStatusACTIVE)
	if err != nil {
		return nil, translateRepoError(err)
	}

	s.logger.Info("integration synced", "integration_id", updated.ID)

	return &clientsgrpc.SyncIntegrationResponse{
		Integration: sqlcIntegrationToProto(updated),
	}, nil
}