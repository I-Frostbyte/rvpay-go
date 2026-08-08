package service

import (
	"time"

	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	clientsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/clientsgrpc"
	commongrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/commongrpc"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func sqlcClientToProto(c sqlc.Client) *clientsgrpc.Client {
	return &clientsgrpc.Client{
		Id:        c.ID.String(),
		Name:      c.ClientName,
		Status:    commongrpc.ClientStatus(commongrpc.ClientStatus_value[string(c.Status)]),
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}

func sqlcPlatformToProto(p sqlc.Platform) *clientsgrpc.Platform {
	capabilities := make([]commongrpc.ProviderCapability, 0, 2)
	if p.OauthCapable {
		capabilities = append(capabilities, commongrpc.ProviderCapability_PROVIDER_CAPABILITY_OAUTH)
	}
	if p.WebhookCapable {
		capabilities = append(capabilities, commongrpc.ProviderCapability_PROVIDER_CAPABILITY_WEBHOOK)
	}
	platformStatus := commongrpc.PlatformStatus_PLATFORM_STATUS_DISABLED
	if p.Enabled {
		platformStatus = commongrpc.PlatformStatus_PLATFORM_STATUS_ACTIVE
	}
	return &clientsgrpc.Platform{
		Id:          p.ID.String(),
		Name:        p.Name,
		DisplayName: p.DisplayName,
		Slug:        p.Slug,
		Status:      platformStatus,
		Capabilities: capabilities,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}

func sqlcIntegrationToProto(i sqlc.Integration) *clientsgrpc.Integration {
	var installedAt, lastSyncAt *timestamppb.Timestamp
	if !i.InstalledAt.IsZero() {
		installedAt = timestamppb.New(i.InstalledAt)
	}
	if i.LastSyncAt.Valid {
		lastSyncAt = timestamppb.New(i.LastSyncAt.Time)
	}
	return &clientsgrpc.Integration{
		Id:                i.ID.String(),
		ClientId:          i.ClientID.String(),
		PlatformId:        i.PlatformID.String(),
		ExternalAccountId: i.ExternalAccountID,
		Status:            commongrpc.IntegrationStatus(commongrpc.IntegrationStatus_value[string(i.Status)]),
		InstalledAt:       installedAt,
		LastSyncAt:        lastSyncAt,
		CreatedAt:         timestamppb.New(i.CreatedAt),
		UpdatedAt:         timestamppb.New(i.UpdatedAt),
	}
}

func protoStatusToSqlcClientStatus(s commongrpc.ClientStatus) sqlc.ClientStatus {
	switch s {
	case commongrpc.ClientStatus_CLIENT_STATUS_ACTIVE:
		return sqlc.ClientStatusACTIVE
	case commongrpc.ClientStatus_CLIENT_STATUS_SUSPENDED:
		return sqlc.ClientStatusSUSPENDED
	case commongrpc.ClientStatus_CLIENT_STATUS_CLOSED:
		return sqlc.ClientStatusCLOSED
	default:
		return sqlc.ClientStatusREGISTERED
	}
}

func protoStatusToSqlcIntegrationStatus(s commongrpc.IntegrationStatus) sqlc.IntegrationStatus {
	switch s {
	case commongrpc.IntegrationStatus_INTEGRATION_STATUS_ACTIVE:
		return sqlc.IntegrationStatusACTIVE
	case commongrpc.IntegrationStatus_INTEGRATION_STATUS_OAUTH_PENDING:
		return sqlc.IntegrationStatusOAUTHPENDING
	case commongrpc.IntegrationStatus_INTEGRATION_STATUS_REVOKED:
		return sqlc.IntegrationStatusREVOKED
	default:
		return sqlc.IntegrationStatusCREATED
	}
}

func parseUUID(id string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}
	return parsed, nil
}

func toTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}