package webhooks

import (
	"context"
	"testing"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/I-Frostbyte/rvpay-go/clients/providers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockWebhookIntegrationRepo is a minimal IntegrationRepo test double
type mockWebhookIntegrationRepo struct {
	integrations map[string]sqlc.Integration
}

func newMockWebhookIntegrationRepo() *mockWebhookIntegrationRepo {
	return &mockWebhookIntegrationRepo{
		integrations: make(map[string]sqlc.Integration),
	}
}

func (m *mockWebhookIntegrationRepo) Create(ctx context.Context, clientID, platformID uuid.UUID, externalAccountID string, status sqlc.IntegrationStatus) (sqlc.Integration, error) {
	return sqlc.Integration{}, nil
}

func (m *mockWebhookIntegrationRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Integration, error) {
	i, ok := m.integrations[id.String()]
	if !ok {
		return sqlc.Integration{}, repo.ErrNotFound
	}
	return i, nil
}

func (m *mockWebhookIntegrationRepo) GetByClientAndPlatform(ctx context.Context, clientID, platformID uuid.UUID) (sqlc.Integration, error) {
	return sqlc.Integration{}, repo.ErrNotFound
}

func (m *mockWebhookIntegrationRepo) GetByExternalAccountID(ctx context.Context, externalAccountID string) (sqlc.Integration, error) {
	return sqlc.Integration{}, repo.ErrNotFound
}

func (m *mockWebhookIntegrationRepo) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	return nil, nil
}

func (m *mockWebhookIntegrationRepo) ListByPlatform(ctx context.Context, platformID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	return nil, nil
}

func (m *mockWebhookIntegrationRepo) ListActiveByClient(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	return nil, nil
}

func (m *mockWebhookIntegrationRepo) CountByClient(ctx context.Context, clientID uuid.UUID) (int64, error) {
	return 0, nil
}

func (m *mockWebhookIntegrationRepo) ExistsByClientAndPlatform(ctx context.Context, clientID, platformID uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockWebhookIntegrationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.IntegrationStatus) (sqlc.Integration, error) {
	return sqlc.Integration{}, nil
}

func (m *mockWebhookIntegrationRepo) UpdateLastSyncAt(ctx context.Context, id uuid.UUID, lastSyncAt pgtype.Timestamptz) (sqlc.Integration, error) {
	return sqlc.Integration{}, nil
}

func (m *mockWebhookIntegrationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// mockWebhookPlatformRepo is a minimal PlatformRepo test double
type mockWebhookPlatformRepo struct {
	platforms map[string]sqlc.Platform
}

func newMockWebhookPlatformRepo() *mockWebhookPlatformRepo {
	return &mockWebhookPlatformRepo{
		platforms: make(map[string]sqlc.Platform),
	}
}

func (m *mockWebhookPlatformRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Platform, error) {
	p, ok := m.platforms[id.String()]
	if !ok {
		return sqlc.Platform{}, repo.ErrNotFound
	}
	return p, nil
}

func (m *mockWebhookPlatformRepo) Create(ctx context.Context, name, displayName, slug string, enabled, oauthCapable, webhookCapable bool) (sqlc.Platform, error) {
	return sqlc.Platform{}, nil
}

func (m *mockWebhookPlatformRepo) GetByName(ctx context.Context, name string) (sqlc.Platform, error) {
	return sqlc.Platform{}, repo.ErrNotFound
}

func (m *mockWebhookPlatformRepo) GetBySlug(ctx context.Context, slug string) (sqlc.Platform, error) {
	return sqlc.Platform{}, repo.ErrNotFound
}

func (m *mockWebhookPlatformRepo) List(ctx context.Context, limit, offset int32) ([]sqlc.Platform, error) {
	return nil, nil
}

func (m *mockWebhookPlatformRepo) ListEnabled(ctx context.Context, limit, offset int32) ([]sqlc.Platform, error) {
	return nil, nil
}

func (m *mockWebhookPlatformRepo) Count(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *mockWebhookPlatformRepo) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return false, nil
}

func (m *mockWebhookPlatformRepo) Update(ctx context.Context, id uuid.UUID, name, displayName, slug string, enabled, oauthCapable, webhookCapable bool) (sqlc.Platform, error) {
	return sqlc.Platform{}, nil
}

func (m *mockWebhookPlatformRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// mockWebhookRepo is a WebhookSubscriptionRepo test double
type mockWebhookRepo struct {
	subscriptions map[string]sqlc.WebhookSubscription
}

func newMockWebhookRepo() *mockWebhookRepo {
	return &mockWebhookRepo{
		subscriptions: make(map[string]sqlc.WebhookSubscription),
	}
}

func (m *mockWebhookRepo) Create(ctx context.Context, integrationID uuid.UUID, endpoint, secret string, status sqlc.WebhookSubscriptionStatus) (sqlc.WebhookSubscription, error) {
	sub := sqlc.WebhookSubscription{
		ID:            uuid.New(),
		IntegrationID: integrationID,
		Endpoint:      endpoint,
		Secret:        secret,
		Status:        status,
	}
	m.subscriptions[sub.ID.String()] = sub
	return sub, nil
}

func (m *mockWebhookRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.WebhookSubscription, error) {
	sub, ok := m.subscriptions[id.String()]
	if !ok {
		return sqlc.WebhookSubscription{}, repo.ErrNotFound
	}
	return sub, nil
}

func (m *mockWebhookRepo) GetByIntegrationIDAndEndpoint(ctx context.Context, integrationID uuid.UUID, endpoint string) (sqlc.WebhookSubscription, error) {
	for _, sub := range m.subscriptions {
		if sub.IntegrationID == integrationID {
			return sub, nil
		}
	}
	return sqlc.WebhookSubscription{}, repo.ErrNotFound
}

func (m *mockWebhookRepo) ListByIntegrationID(ctx context.Context, integrationID uuid.UUID, limit, offset int32) ([]sqlc.WebhookSubscription, error) {
	return nil, nil
}

func (m *mockWebhookRepo) ListActiveByIntegrationID(ctx context.Context, integrationID uuid.UUID) ([]sqlc.WebhookSubscription, error) {
	return nil, nil
}

func (m *mockWebhookRepo) CountByIntegrationID(ctx context.Context, integrationID uuid.UUID) (int64, error) {
	return 0, nil
}

func (m *mockWebhookRepo) Exists(ctx context.Context, integrationID uuid.UUID, endpoint string) (bool, error) {
	return false, nil
}

func (m *mockWebhookRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.WebhookSubscriptionStatus) (sqlc.WebhookSubscription, error) {
	sub, ok := m.subscriptions[id.String()]
	if !ok {
		return sqlc.WebhookSubscription{}, repo.ErrNotFound
	}
	sub.Status = status
	m.subscriptions[id.String()] = sub
	return sub, nil
}

func (m *mockWebhookRepo) UpdateLastDelivery(ctx context.Context, id uuid.UUID, lastDelivery pgtype.Timestamptz) (sqlc.WebhookSubscription, error) {
	sub, ok := m.subscriptions[id.String()]
	if !ok {
		return sqlc.WebhookSubscription{}, repo.ErrNotFound
	}
	sub.LastDelivery = lastDelivery
	m.subscriptions[id.String()] = sub
	return sub, nil
}

func (m *mockWebhookRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := m.subscriptions[id.String()]; !ok {
		return repo.ErrNotFound
	}
	delete(m.subscriptions, id.String())
	return nil
}

func TestProcessWebhookUnknownProvider(t *testing.T) {
	t.Parallel()

	registry := providers.NewProviderRegistry()

	svc := NewService(
		newMockWebhookIntegrationRepo(),
		newMockWebhookRepo(),
		newMockWebhookPlatformRepo(),
		registry,
		zerolog.Nop(),
	)

	err := svc.ProcessWebhook(context.Background(), "unknown", map[string]string{}, []byte(`{}`))
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("status code = %s, want %s", status.Code(err), codes.FailedPrecondition)
	}
}

func TestProcessWebhookInvalidSignature(t *testing.T) {
	t.Parallel()

	registry := providers.NewProviderRegistry()
	registry.Register(providers.NewHighLevelProvider("test-client", "test-secret", "https://example.com/callback", "test-webhook-secret"))

	svc := NewService(
		newMockWebhookIntegrationRepo(),
		newMockWebhookRepo(),
		newMockWebhookPlatformRepo(),
		registry,
		zerolog.Nop(),
	)

	headers := map[string]string{}
	err := svc.ProcessWebhook(context.Background(), "highlevel", headers, []byte(`{"eventId":"test"}`))
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("status code = %s, want %s", status.Code(err), codes.InvalidArgument)
	}
}

func TestRegisterWebhookIntegrationNotFound(t *testing.T) {
	t.Parallel()

	registry := providers.NewProviderRegistry()
	registry.Register(providers.NewHighLevelProvider("test-client", "test-secret", "https://example.com/callback", "test-webhook-secret"))

	svc := NewService(
		newMockWebhookIntegrationRepo(),
		newMockWebhookRepo(),
		newMockWebhookPlatformRepo(),
		registry,
		zerolog.Nop(),
	)

	err := svc.RegisterWebhook(context.Background(), uuid.New(), "https://example.com/callback")
	if status.Code(err) != codes.NotFound {
		t.Fatalf("status code = %s, want %s", status.Code(err), codes.NotFound)
	}
}

func TestUnregisterWebhookNotFound(t *testing.T) {
	t.Parallel()

	registry := providers.NewProviderRegistry()
	registry.Register(providers.NewHighLevelProvider("test-client", "test-secret", "https://example.com/callback", "test-webhook-secret"))

	svc := NewService(
		newMockWebhookIntegrationRepo(),
		newMockWebhookRepo(),
		newMockWebhookPlatformRepo(),
		registry,
		zerolog.Nop(),
	)

	err := svc.UnregisterWebhook(context.Background(), uuid.New())
	if status.Code(err) != codes.NotFound {
		t.Fatalf("status code = %s, want %s", status.Code(err), codes.NotFound)
	}
}
