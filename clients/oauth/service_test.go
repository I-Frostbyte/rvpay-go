package oauth

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/I-Frostbyte/rvpay-go/clients/providers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockClientRepo is a test double for ClientRepo
type mockClientRepo struct {
	clients map[string]sqlc.Client
}

func newMockClientRepo() *mockClientRepo {
	return &mockClientRepo{
		clients: make(map[string]sqlc.Client),
	}
}

func (m *mockClientRepo) Create(ctx context.Context, name string, status sqlc.ClientStatus) (sqlc.Client, error) {
	client := sqlc.Client{
		ID:         uuid.New(),
		ClientName: name,
		Status:     status,
	}
	m.clients[client.ID.String()] = client
	return client, nil
}

func (m *mockClientRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Client, error) {
	client, ok := m.clients[id.String()]
	if !ok {
		return sqlc.Client{}, repo.ErrNotFound
	}
	return client, nil
}

func (m *mockClientRepo) GetByName(ctx context.Context, name string) (sqlc.Client, error) {
	for _, client := range m.clients {
		if client.ClientName == name {
			return client, nil
		}
	}
	return sqlc.Client{}, repo.ErrNotFound
}

func (m *mockClientRepo) List(ctx context.Context, limit, offset int32) ([]sqlc.Client, error) {
	clients := make([]sqlc.Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	return clients, nil
}

func (m *mockClientRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(m.clients)), nil
}

func (m *mockClientRepo) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	_, ok := m.clients[id.String()]
	return ok, nil
}

func (m *mockClientRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.ClientStatus) (sqlc.Client, error) {
	client, ok := m.clients[id.String()]
	if !ok {
		return sqlc.Client{}, repo.ErrNotFound
	}
	client.Status = status
	m.clients[id.String()] = client
	return client, nil
}

func (m *mockClientRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := m.clients[id.String()]; !ok {
		return repo.ErrNotFound
	}
	delete(m.clients, id.String())
	return nil
}

func (m *mockClientRepo) ListActive(ctx context.Context, limit, offset int32) ([]sqlc.Client, error) {
	clients := make([]sqlc.Client, 0)
	for _, client := range m.clients {
		if client.Status == sqlc.ClientStatusACTIVE {
			clients = append(clients, client)
		}
	}
	return clients, nil
}

// mockPlatformRepo is a test double for PlatformRepo
type mockPlatformRepo struct {
	platforms map[string]sqlc.Platform
}

func newMockPlatformRepo() *mockPlatformRepo {
	return &mockPlatformRepo{
		platforms: make(map[string]sqlc.Platform),
	}
}

func (m *mockPlatformRepo) Create(ctx context.Context, name, displayName, slug string, enabled, oauthCapable, webhookCapable bool) (sqlc.Platform, error) {
	platform := sqlc.Platform{
		ID:             uuid.New(),
		Name:           name,
		DisplayName:    displayName,
		Slug:           slug,
		Enabled:        enabled,
		OauthCapable:   oauthCapable,
		WebhookCapable: webhookCapable,
	}
	m.platforms[platform.ID.String()] = platform
	return platform, nil
}

func (m *mockPlatformRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Platform, error) {
	p, ok := m.platforms[id.String()]
	if !ok {
		return sqlc.Platform{}, repo.ErrNotFound
	}
	return p, nil
}

func (m *mockPlatformRepo) GetByName(ctx context.Context, name string) (sqlc.Platform, error) {
	for _, p := range m.platforms {
		if p.Name == name {
			return p, nil
		}
	}
	return sqlc.Platform{}, repo.ErrNotFound
}

func (m *mockPlatformRepo) GetBySlug(ctx context.Context, slug string) (sqlc.Platform, error) {
	for _, p := range m.platforms {
		if p.Slug == slug {
			return p, nil
		}
	}
	return sqlc.Platform{}, repo.ErrNotFound
}

func (m *mockPlatformRepo) List(ctx context.Context, limit, offset int32) ([]sqlc.Platform, error) {
	platforms := make([]sqlc.Platform, 0, len(m.platforms))
	for _, p := range m.platforms {
		platforms = append(platforms, p)
	}
	return platforms, nil
}

func (m *mockPlatformRepo) ListEnabled(ctx context.Context, limit, offset int32) ([]sqlc.Platform, error) {
	platforms := make([]sqlc.Platform, 0)
	for _, p := range m.platforms {
		if p.Enabled {
			platforms = append(platforms, p)
		}
	}
	return platforms, nil
}

func (m *mockPlatformRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(m.platforms)), nil
}

func (m *mockPlatformRepo) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	for _, p := range m.platforms {
		if p.Slug == slug {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockPlatformRepo) Update(ctx context.Context, id uuid.UUID, name, displayName, slug string, enabled, oauthCapable, webhookCapable bool) (sqlc.Platform, error) {
	p, ok := m.platforms[id.String()]
	if !ok {
		return sqlc.Platform{}, repo.ErrNotFound
	}
	p.Name = name
	p.DisplayName = displayName
	p.Slug = slug
	p.Enabled = enabled
	p.OauthCapable = oauthCapable
	p.WebhookCapable = webhookCapable
	m.platforms[id.String()] = p
	return p, nil
}

func (m *mockPlatformRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := m.platforms[id.String()]; !ok {
		return repo.ErrNotFound
	}
	delete(m.platforms, id.String())
	return nil
}

// mockIntegrationRepo is a test double for IntegrationRepo
type mockIntegrationRepo struct {
	integrations map[string]sqlc.Integration
}

func newMockIntegrationRepo() *mockIntegrationRepo {
	return &mockIntegrationRepo{
		integrations: make(map[string]sqlc.Integration),
	}
}

func (m *mockIntegrationRepo) Create(ctx context.Context, clientID, platformID uuid.UUID, externalAccountID string, status sqlc.IntegrationStatus) (sqlc.Integration, error) {
	integration := sqlc.Integration{
		ID:                uuid.New(),
		ClientID:          clientID,
		PlatformID:        platformID,
		ExternalAccountID: externalAccountID,
		Status:            status,
		InstalledAt:       time.Now(),
	}
	m.integrations[integration.ID.String()] = integration
	return integration, nil
}

func (m *mockIntegrationRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Integration, error) {
	integration, ok := m.integrations[id.String()]
	if !ok {
		return sqlc.Integration{}, repo.ErrNotFound
	}
	return integration, nil
}

func (m *mockIntegrationRepo) GetByClientAndPlatform(ctx context.Context, clientID, platformID uuid.UUID) (sqlc.Integration, error) {
	for _, integration := range m.integrations {
		if integration.ClientID == clientID && integration.PlatformID == platformID {
			return integration, nil
		}
	}
	return sqlc.Integration{}, repo.ErrNotFound
}

func (m *mockIntegrationRepo) GetByExternalAccountID(ctx context.Context, externalAccountID string) (sqlc.Integration, error) {
	for _, integration := range m.integrations {
		if integration.ExternalAccountID == externalAccountID {
			return integration, nil
		}
	}
	return sqlc.Integration{}, repo.ErrNotFound
}

func (m *mockIntegrationRepo) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	integrations := make([]sqlc.Integration, 0)
	for _, integration := range m.integrations {
		if integration.ClientID == clientID {
			integrations = append(integrations, integration)
		}
	}
	return integrations, nil
}

func (m *mockIntegrationRepo) ListByPlatform(ctx context.Context, platformID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	integrations := make([]sqlc.Integration, 0)
	for _, integration := range m.integrations {
		if integration.PlatformID == platformID {
			integrations = append(integrations, integration)
		}
	}
	return integrations, nil
}

func (m *mockIntegrationRepo) ListActiveByClient(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	integrations := make([]sqlc.Integration, 0)
	for _, integration := range m.integrations {
		if integration.ClientID == clientID && integration.Status == sqlc.IntegrationStatusACTIVE {
			integrations = append(integrations, integration)
		}
	}
	return integrations, nil
}

func (m *mockIntegrationRepo) CountByClient(ctx context.Context, clientID uuid.UUID) (int64, error) {
	count := int64(0)
	for _, integration := range m.integrations {
		if integration.ClientID == clientID {
			count++
		}
	}
	return count, nil
}

func (m *mockIntegrationRepo) ExistsByClientAndPlatform(ctx context.Context, clientID, platformID uuid.UUID) (bool, error) {
	for _, integration := range m.integrations {
		if integration.ClientID == clientID && integration.PlatformID == platformID {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockIntegrationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.IntegrationStatus) (sqlc.Integration, error) {
	integration, ok := m.integrations[id.String()]
	if !ok {
		return sqlc.Integration{}, repo.ErrNotFound
	}
	integration.Status = status
	m.integrations[id.String()] = integration
	return integration, nil
}

func (m *mockIntegrationRepo) UpdateLastSyncAt(ctx context.Context, id uuid.UUID, lastSyncAt pgtype.Timestamptz) (sqlc.Integration, error) {
	integration, ok := m.integrations[id.String()]
	if !ok {
		return sqlc.Integration{}, repo.ErrNotFound
	}
	integration.LastSyncAt = lastSyncAt
	m.integrations[id.String()] = integration
	return integration, nil
}

func (m *mockIntegrationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := m.integrations[id.String()]; !ok {
		return repo.ErrNotFound
	}
	delete(m.integrations, id.String())
	return nil
}

func (m *mockIntegrationRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(m.integrations)), nil
}

func (m *mockIntegrationRepo) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	_, ok := m.integrations[id.String()]
	return ok, nil
}

// mockOAuthTokenRepo is a test double for OAuthTokenRepo
type mockOAuthTokenRepo struct {
	tokens map[string]sqlc.OauthToken
}

func newMockOAuthTokenRepo() *mockOAuthTokenRepo {
	return &mockOAuthTokenRepo{
		tokens: make(map[string]sqlc.OauthToken),
	}
}

func (m *mockOAuthTokenRepo) Create(ctx context.Context, integrationID uuid.UUID, accessToken, refreshToken string, expiresAt time.Time, scope, tokenType string) (sqlc.OauthToken, error) {
	token := sqlc.OauthToken{
		ID:            uuid.New(),
		IntegrationID: integrationID,
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		ExpiresAt:     expiresAt,
		Scope:         scope,
		TokenType:     tokenType,
	}
	m.tokens[token.ID.String()] = token
	return token, nil
}

func (m *mockOAuthTokenRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.OauthToken, error) {
	token, ok := m.tokens[id.String()]
	if !ok {
		return sqlc.OauthToken{}, repo.ErrNotFound
	}
	return token, nil
}

func (m *mockOAuthTokenRepo) GetByIntegrationID(ctx context.Context, integrationID uuid.UUID) (sqlc.OauthToken, error) {
	for _, token := range m.tokens {
		if token.IntegrationID == integrationID {
			return token, nil
		}
	}
	return sqlc.OauthToken{}, repo.ErrNotFound
}

func (m *mockOAuthTokenRepo) ExistsByIntegrationID(ctx context.Context, integrationID uuid.UUID) (bool, error) {
	for _, token := range m.tokens {
		if token.IntegrationID == integrationID {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockOAuthTokenRepo) Update(ctx context.Context, id uuid.UUID, accessToken, refreshToken string, expiresAt time.Time, scope, tokenType string) (sqlc.OauthToken, error) {
	token, ok := m.tokens[id.String()]
	if !ok {
		return sqlc.OauthToken{}, repo.ErrNotFound
	}
	token.AccessToken = accessToken
	token.RefreshToken = refreshToken
	token.ExpiresAt = expiresAt
	token.Scope = scope
	token.TokenType = tokenType
	m.tokens[id.String()] = token
	return token, nil
}

func (m *mockOAuthTokenRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := m.tokens[id.String()]; !ok {
		return repo.ErrNotFound
	}
	delete(m.tokens, id.String())
	return nil
}

func (m *mockOAuthTokenRepo) DeleteByIntegrationID(ctx context.Context, integrationID uuid.UUID) error {
	for id, token := range m.tokens {
		if token.IntegrationID == integrationID {
			delete(m.tokens, id)
		}
	}
	return nil
}

func TestAuthorizationURL(t *testing.T) {
	t.Parallel()

	platformRepo := newMockPlatformRepo()
	platformID := uuid.New()
	platformRepo.platforms[platformID.String()] = sqlc.Platform{
		ID:      platformID,
		Name:    "HighLevel",
		Slug:    "highlevel",
		Enabled: true,
	}

	registry := providers.NewProviderRegistry()
	registry.Register(providers.NewHighLevelProvider("test-client", "test-secret", "https://example.com/callback", "test-webhook-secret"))

	svc := NewService(
		newMockIntegrationRepo(),
		newMockOAuthTokenRepo(),
		newMockClientRepo(),
		platformRepo,
		registry,
		"https://example.com/callback",
		zerolog.Nop(),
	)

	authURL, err := svc.AuthorizationURL(context.Background(), uuid.New(), platformID, "test-state")
	if err != nil {
		t.Fatalf("AuthorizationURL failed: %v", err)
	}
	if authURL == "" {
		t.Fatal("authorization URL should not be empty")
	}

	// SECURITY REGRESSION TEST (SEC-02): the redirect_uri in the generated
	// authorization URL must be the configured value, never a hard-coded
	// fallback.
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("authorization URL unparseable: %v", err)
	}
	if got := parsed.Query().Get("redirect_uri"); got != "https://example.com/callback" {
		t.Fatalf("redirect_uri = %q, want configured %q", got, "https://example.com/callback")
	}
}

func TestAuthorizationURLDisabledPlatform(t *testing.T) {
	t.Parallel()

	platformRepo := newMockPlatformRepo()
	platformID := uuid.New()
	platformRepo.platforms[platformID.String()] = sqlc.Platform{
		ID:      platformID,
		Name:    "HighLevel",
		Slug:    "highlevel",
		Enabled: false,
	}

	registry := providers.NewProviderRegistry()
	registry.Register(providers.NewHighLevelProvider("test-client", "test-secret", "https://example.com/callback", "test-webhook-secret"))

	svc := NewService(
		newMockIntegrationRepo(),
		newMockOAuthTokenRepo(),
		newMockClientRepo(),
		platformRepo,
		registry,
		"https://example.com/callback",
		zerolog.Nop(),
	)

	_, err := svc.AuthorizationURL(context.Background(), uuid.New(), platformID, "test-state")
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("status code = %s, want %s", status.Code(err), codes.FailedPrecondition)
	}
}

func TestAuthorizationURLUnknownProvider(t *testing.T) {
	t.Parallel()

	platformRepo := newMockPlatformRepo()
	platformID := uuid.New()
	platformRepo.platforms[platformID.String()] = sqlc.Platform{
		ID:      platformID,
		Name:    "Unknown",
		Slug:    "unknown",
		Enabled: true,
	}

	registry := providers.NewProviderRegistry()

	svc := NewService(
		newMockIntegrationRepo(),
		newMockOAuthTokenRepo(),
		newMockClientRepo(),
		platformRepo,
		registry,
		"https://example.com/callback",
		zerolog.Nop(),
	)

	_, err := svc.AuthorizationURL(context.Background(), uuid.New(), platformID, "test-state")
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("status code = %s, want %s", status.Code(err), codes.FailedPrecondition)
	}
}
