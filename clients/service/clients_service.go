package service

import (
	"context"
	"errors"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	clientsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/clientsgrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ClientsServiceImpl struct {
	clientsRepo repo.ClientRepo
	logger      Logger
}

func NewClientsServiceImpl(clientsRepo repo.ClientRepo, logger Logger) *ClientsServiceImpl {
	return &ClientsServiceImpl{
		clientsRepo: clientsRepo,
		logger:      logger,
	}
}

func (s *ClientsServiceImpl) CreateClient(ctx context.Context, req *clientsgrpc.CreateClientRequest) (*clientsgrpc.CreateClientResponse, error) {
	if req == nil || req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "client name is required")
	}

	_, err := s.clientsRepo.GetByName(ctx, req.GetName())
	if err == nil {
		return nil, ErrClientAlreadyExists
	}
	if !errors.Is(err, repo.ErrNotFound) {
		return nil, translateRepoError(err)
	}

	client, err := s.clientsRepo.Create(ctx, req.GetName(), sqlc.ClientStatusREGISTERED)
	if err != nil {
		return nil, translateRepoError(err)
	}

	s.logger.Info("client created", "client_id", client.ID, "name", client.ClientName)

	return &clientsgrpc.CreateClientResponse{
		Client: sqlcClientToProto(client),
	}, nil
}

func (s *ClientsServiceImpl) GetClient(ctx context.Context, req *clientsgrpc.GetClientRequest) (*clientsgrpc.GetClientResponse, error) {
	id, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	client, err := s.clientsRepo.GetByID(ctx, id)
	if err == repo.ErrNotFound {
		return nil, ErrClientNotFound
	}
	if err != nil {
		return nil, translateRepoError(err)
	}

	return &clientsgrpc.GetClientResponse{
		Client: sqlcClientToProto(client),
	}, nil
}

func (s *ClientsServiceImpl) ListClients(ctx context.Context, req *clientsgrpc.ListClientsRequest) (*clientsgrpc.ListClientsResponse, error) {
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

	clients, err := s.clientsRepo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, translateRepoError(err)
	}

	protoClients := make([]*clientsgrpc.Client, 0, len(clients))
	for _, c := range clients {
		protoClients = append(protoClients, sqlcClientToProto(c))
	}

	return &clientsgrpc.ListClientsResponse{
		Clients: protoClients,
	}, nil
}

func (s *ClientsServiceImpl) UpdateClient(ctx context.Context, req *clientsgrpc.UpdateClientRequest) (*clientsgrpc.UpdateClientResponse, error) {
	if req == nil || req.GetId() == "" || req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "client id and name are required")
	}

	id, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	client, err := s.clientsRepo.GetByID(ctx, id)
	if err == repo.ErrNotFound {
		return nil, ErrClientNotFound
	}
	if err != nil {
		return nil, translateRepoError(err)
	}

	updated, err := s.clientsRepo.UpdateStatus(ctx, id, client.Status)
	if err != nil {
		return nil, translateRepoError(err)
	}

	s.logger.Info("client updated", "client_id", updated.ID, "name", updated.ClientName)

	return &clientsgrpc.UpdateClientResponse{
		Client: sqlcClientToProto(updated),
	}, nil
}

func (s *ClientsServiceImpl) DeleteClient(ctx context.Context, req *clientsgrpc.DeleteClientRequest) (*clientsgrpc.DeleteClientResponse, error) {
	id, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	client, err := s.clientsRepo.GetByID(ctx, id)
	if err == repo.ErrNotFound {
		return nil, ErrClientNotFound
	}
	if err != nil {
		return nil, translateRepoError(err)
	}

	if client.Status == sqlc.ClientStatusACTIVE {
		return nil, ErrClientHasIntegrations
	}

	err = s.clientsRepo.Delete(ctx, id)
	if err != nil {
		return nil, translateRepoError(err)
	}

	s.logger.Info("client deleted", "client_id", id)

	return &clientsgrpc.DeleteClientResponse{
		Id: id.String(),
	}, nil
}

func (s *ClientsServiceImpl) ActivateClient(ctx context.Context, req *clientsgrpc.ActivateClientRequest) (*clientsgrpc.ActivateClientResponse, error) {
	id, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	client, err := s.clientsRepo.GetByID(ctx, id)
	if err == repo.ErrNotFound {
		return nil, ErrClientNotFound
	}
	if err != nil {
		return nil, translateRepoError(err)
	}

	if client.Status == sqlc.ClientStatusACTIVE {
		return &clientsgrpc.ActivateClientResponse{
			Client: sqlcClientToProto(client),
		}, nil
	}

	updated, err := s.clientsRepo.UpdateStatus(ctx, id, sqlc.ClientStatusACTIVE)
	if err != nil {
		return nil, translateRepoError(err)
	}

	s.logger.Info("client activated", "client_id", updated.ID)

	return &clientsgrpc.ActivateClientResponse{
		Client: sqlcClientToProto(updated),
	}, nil
}

func (s *ClientsServiceImpl) DeactivateClient(ctx context.Context, req *clientsgrpc.DeactivateClientRequest) (*clientsgrpc.DeactivateClientResponse, error) {
	id, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	client, err := s.clientsRepo.GetByID(ctx, id)
	if err == repo.ErrNotFound {
		return nil, ErrClientNotFound
	}
	if err != nil {
		return nil, translateRepoError(err)
	}

	if client.Status == sqlc.ClientStatusCLOSED {
		return &clientsgrpc.DeactivateClientResponse{
			Client: sqlcClientToProto(client),
		}, nil
	}

	updated, err := s.clientsRepo.UpdateStatus(ctx, id, sqlc.ClientStatusCLOSED)
	if err != nil {
		return nil, translateRepoError(err)
	}

	s.logger.Info("client deactivated", "client_id", updated.ID)

	return &clientsgrpc.DeactivateClientResponse{
		Client: sqlcClientToProto(updated),
	}, nil
}

type Logger interface {
	Info(msg string, args ...interface{})
}