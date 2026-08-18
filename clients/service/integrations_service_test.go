package service

import (
	"context"
	"testing"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/repo/mocks"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	clientsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/clientsgrpc"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"go.uber.org/mock/gomock"
)

// newTestIntegrationsService builds an IntegrationsServiceImpl wired to
// generated mocks. It returns the service and the mocks for assertions.
func newTestIntegrationsService(t *testing.T) (*IntegrationsServiceImpl, *mocks.MockIntegrationRepo, *mocks.MockClientRepo, *mocks.MockPlatformRepo) {
	t.Helper()

	ctrl := gomock.NewController(t)
	integrationRepo := mocks.NewMockIntegrationRepo(ctrl)
	clientRepo := mocks.NewMockClientRepo(ctrl)
	platformRepo := mocks.NewMockPlatformRepo(ctrl)
	oauthRepo := mocks.NewMockOAuthTokenRepo(ctrl)
	webhookRepo := mocks.NewMockWebhookSubscriptionRepo(ctrl)

	svc := NewIntegrationsServiceImpl(
		integrationRepo,
		clientRepo,
		platformRepo,
		oauthRepo,
		webhookRepo,
		zerolog.Nop(),
	)

	return svc, integrationRepo, clientRepo, platformRepo
}

func TestInstallIntegration_StoresExternalAccountID(t *testing.T) {
	t.Parallel()

	svc, integrationRepo, clientRepo, platformRepo := newTestIntegrationsService(t)

	clientID := uuid.New()
	platformID := uuid.New()
	integrationID := uuid.New()

	clientRepo.EXPECT().GetByID(gomock.Any(), clientID).Return(sqlc.Client{
		ID:     clientID,
		Status: sqlc.ClientStatusACTIVE,
	}, nil)
	platformRepo.EXPECT().GetByID(gomock.Any(), platformID).Return(sqlc.Platform{
		ID:      platformID,
		Enabled: true,
	}, nil)
	integrationRepo.EXPECT().GetByClientAndPlatform(gomock.Any(), clientID, platformID).Return(sqlc.Integration{}, repo.ErrNotFound)
	integrationRepo.EXPECT().Create(gomock.Any(), clientID, platformID, "GHL_LOCATION_A", sqlc.IntegrationStatusCREATED).Return(sqlc.Integration{
		ID:                integrationID,
		ClientID:          clientID,
		PlatformID:        platformID,
		ExternalAccountID: "GHL_LOCATION_A",
		Status:            sqlc.IntegrationStatusCREATED,
	}, nil)

	resp, err := svc.InstallIntegration(context.Background(), &clientsgrpc.InstallIntegrationRequest{
		ClientId:          clientID.String(),
		PlatformId:        platformID.String(),
		ExternalAccountId: "GHL_LOCATION_A",
	})
	if err != nil {
		t.Fatalf("InstallIntegration failed: %v", err)
	}

	if resp.Integration.ExternalAccountId != "GHL_LOCATION_A" {
		t.Fatalf("ExternalAccountId = %q, want GHL_LOCATION_A", resp.Integration.ExternalAccountId)
	}
	if resp.Integration.ClientId != clientID.String() {
		t.Fatalf("ClientId = %q, want %q", resp.Integration.ClientId, clientID.String())
	}
	if resp.Integration.PlatformId != platformID.String() {
		t.Fatalf("PlatformId = %q, want %q", resp.Integration.PlatformId, platformID.String())
	}
}

func TestInstallIntegration_WithoutExternalAccountID_BackwardsCompatible(t *testing.T) {
	t.Parallel()

	svc, integrationRepo, clientRepo, platformRepo := newTestIntegrationsService(t)

	clientID := uuid.New()
	platformID := uuid.New()
	integrationID := uuid.New()

	clientRepo.EXPECT().GetByID(gomock.Any(), clientID).Return(sqlc.Client{
		ID:     clientID,
		Status: sqlc.ClientStatusACTIVE,
	}, nil)
	platformRepo.EXPECT().GetByID(gomock.Any(), platformID).Return(sqlc.Platform{
		ID:      platformID,
		Enabled: true,
	}, nil)
	integrationRepo.EXPECT().GetByClientAndPlatform(gomock.Any(), clientID, platformID).Return(sqlc.Integration{}, repo.ErrNotFound)
	// When external_account_id is omitted, the integration is created with an
	// empty external account ID (backwards-compatible behavior).
	integrationRepo.EXPECT().Create(gomock.Any(), clientID, platformID, "", sqlc.IntegrationStatusCREATED).Return(sqlc.Integration{
		ID:         integrationID,
		ClientID:   clientID,
		PlatformID: platformID,
		Status:     sqlc.IntegrationStatusCREATED,
	}, nil)

	resp, err := svc.InstallIntegration(context.Background(), &clientsgrpc.InstallIntegrationRequest{
		ClientId:   clientID.String(),
		PlatformId: platformID.String(),
	})
	if err != nil {
		t.Fatalf("InstallIntegration failed: %v", err)
	}

	if resp.Integration.ExternalAccountId != "" {
		t.Fatalf("ExternalAccountId = %q, want empty (backwards-compatible)", resp.Integration.ExternalAccountId)
	}
}
