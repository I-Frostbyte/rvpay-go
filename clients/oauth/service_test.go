package oauth

import (
	"context"
	"net/http"
	"net/http/httptest"
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

// mockOAuthStateRepo is a test double for OAuthStateRepo
type mockOAuthStateRepo struct {
	states map[string]sqlc.OauthState
}

func newMockOAuthStateRepo() *mockOAuthStateRepo {
	return &mockOAuthStateRepo{
		states: make(map[string]sqlc.OauthState),
	}
}

func (m *mockOAuthStateRepo) Create(ctx context.Context, state string, clientID, platformID uuid.UUID, expiresAt time.Time) (sqlc.OauthState, error) {
	record := sqlc.OauthState{
		ID:         uuid.New(),
		State:      state,
		ClientID:   clientID,
		PlatformID: platformID,
		ExpiresAt:  expiresAt,
	}
	m.states[state] = record
	return record, nil
}

func (m *mockOAuthStateRepo) GetByState(ctx context.Context, state string) (sqlc.OauthState, error) {
	record, ok := m.states[state]
	if !ok {
		return sqlc.OauthState{}, repo.ErrNotFound
	}
	return record, nil
}

func (m *mockOAuthStateRepo) Consume(ctx context.Context, state string) (sqlc.OauthState, error) {
	record, ok := m.states[state]
	if !ok {
		return sqlc.OauthState{}, repo.ErrNotFound
	}
	if record.ConsumedAt.Valid {
		return sqlc.OauthState{}, repo.ErrNotFound
	}
	if time.Now().After(record.ExpiresAt) {
		return sqlc.OauthState{}, repo.ErrNotFound
	}
	record.ConsumedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	m.states[state] = record
	return record, nil
}

func (m *mockOAuthStateRepo) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

// mockPaymentProviderConfigRepo is a test double for PaymentProviderConfigRepo
type mockPaymentProviderConfigRepo struct {
	configs map[string]sqlc.PaymentProviderConfig
}

func newMockPaymentProviderConfigRepo() *mockPaymentProviderConfigRepo {
	return &mockPaymentProviderConfigRepo{
		configs: make(map[string]sqlc.PaymentProviderConfig),
	}
}

func (m *mockPaymentProviderConfigRepo) Create(ctx context.Context, integrationID uuid.UUID, providerName, providerDescription, providerImageURL, locationID, queryURL, paymentsURL string, supportsSubscriptionSchedule bool, providerAPIKey string) (sqlc.PaymentProviderConfig, error) {
	config := sqlc.PaymentProviderConfig{
		ID:                           uuid.New(),
		IntegrationID:                integrationID,
		ProviderName:                 providerName,
		ProviderDescription:          providerDescription,
		ProviderImageUrl:             providerImageURL,
		LocationID:                   locationID,
		QueryUrl:                     queryURL,
		PaymentsUrl:                  paymentsURL,
		SupportsSubscriptionSchedule: supportsSubscriptionSchedule,
		ProviderApiKey:               providerAPIKey,
	}
	m.configs[config.ID.String()] = config
	return config, nil
}

func (m *mockPaymentProviderConfigRepo) GetByIntegrationID(ctx context.Context, integrationID uuid.UUID) (sqlc.PaymentProviderConfig, error) {
	for _, config := range m.configs {
		if config.IntegrationID == integrationID {
			return config, nil
		}
	}
	return sqlc.PaymentProviderConfig{}, repo.ErrNotFound
}

func (m *mockPaymentProviderConfigRepo) GetByLocationID(ctx context.Context, locationID string) (sqlc.PaymentProviderConfig, error) {
	for _, config := range m.configs {
		if config.LocationID == locationID {
			return config, nil
		}
	}
	return sqlc.PaymentProviderConfig{}, repo.ErrNotFound
}

func (m *mockPaymentProviderConfigRepo) GetByAPIKey(ctx context.Context, apiKey string) (sqlc.PaymentProviderConfig, error) {
	for _, config := range m.configs {
		if config.ProviderApiKey == apiKey {
			return config, nil
		}
	}
	return sqlc.PaymentProviderConfig{}, repo.ErrNotFound
}

func (m *mockPaymentProviderConfigRepo) Update(ctx context.Context, integrationID uuid.UUID, providerName, providerDescription, providerImageURL, locationID, queryURL, paymentsURL string, supportsSubscriptionSchedule bool, providerAPIKey string) (sqlc.PaymentProviderConfig, error) {
	for id, config := range m.configs {
		if config.IntegrationID == integrationID {
			config.ProviderName = providerName
			config.ProviderDescription = providerDescription
			config.ProviderImageUrl = providerImageURL
			config.LocationID = locationID
			config.QueryUrl = queryURL
			config.PaymentsUrl = paymentsURL
			config.SupportsSubscriptionSchedule = supportsSubscriptionSchedule
			config.ProviderApiKey = providerAPIKey
			m.configs[id] = config
			return config, nil
		}
	}
	return sqlc.PaymentProviderConfig{}, repo.ErrNotFound
}

func (m *mockPaymentProviderConfigRepo) Delete(ctx context.Context, integrationID uuid.UUID) error {
	for id, config := range m.configs {
		if config.IntegrationID == integrationID {
			delete(m.configs, id)
			return nil
		}
	}
	return repo.ErrNotFound
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
	registry.Register(providers.NewHighLevelProvider("test-client", "test-secret", "https://example.com/callback", "test-webhook-secret", nil))

	svc := NewService(
		newMockIntegrationRepo(),
		newMockOAuthTokenRepo(),
		newMockClientRepo(),
		platformRepo,
		newMockOAuthStateRepo(),
		newMockPaymentProviderConfigRepo(),
		registry,
		"https://example.com/callback",
		ProviderConfigSettings{},
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
	registry.Register(providers.NewHighLevelProvider("test-client", "test-secret", "https://example.com/callback", "test-webhook-secret", nil))

	svc := NewService(
		newMockIntegrationRepo(),
		newMockOAuthTokenRepo(),
		newMockClientRepo(),
		platformRepo,
		newMockOAuthStateRepo(),
		newMockPaymentProviderConfigRepo(),
		registry,
		"https://example.com/callback",
		ProviderConfigSettings{},
		zerolog.Nop(),
	)

	_, err := svc.AuthorizationURL(context.Background(), uuid.New(), platformID, "test-state")
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("status code = %s, want %s", status.Code(err), codes.FailedPrecondition)
	}
}

func TestBeginAuthorization(t *testing.T) {
	t.Parallel()

	clientRepo := newMockClientRepo()
	platformRepo := newMockPlatformRepo()
	stateRepo := newMockOAuthStateRepo()

	clientID := uuid.New()
	platformID := uuid.New()
	clientRepo.clients[clientID.String()] = sqlc.Client{ID: clientID, Status: sqlc.ClientStatusACTIVE}
	platformRepo.platforms[platformID.String()] = sqlc.Platform{ID: platformID, Name: "HighLevel", Slug: "highlevel", Enabled: true}

	registry := providers.NewProviderRegistry()
	registry.Register(providers.NewHighLevelProvider("test-client", "test-secret", "https://example.com/callback", "", nil))

	svc := NewService(
		newMockIntegrationRepo(),
		newMockOAuthTokenRepo(),
		clientRepo,
		platformRepo,
		stateRepo,
		newMockPaymentProviderConfigRepo(),
		registry,
		"https://example.com/callback",
		ProviderConfigSettings{},
		zerolog.Nop(),
	)

	authURL, state, err := svc.BeginAuthorization(context.Background(), clientID, platformID)
	if err != nil {
		t.Fatalf("BeginAuthorization failed: %v", err)
	}
	if authURL == "" {
		t.Fatal("authorization URL should not be empty")
	}
	if state == "" {
		t.Fatal("state should not be empty")
	}

	// The state must be persisted so the callback can recover context.
	if _, ok := stateRepo.states[state]; !ok {
		t.Fatal("state was not persisted")
	}

	// The redirect_uri in the auth URL must be the configured value.
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("authorization URL unparseable: %v", err)
	}
	if got := parsed.Query().Get("redirect_uri"); got != "https://example.com/callback" {
		t.Fatalf("redirect_uri = %q, want configured %q", got, "https://example.com/callback")
	}
}

func TestBeginAuthorizationInactiveClient(t *testing.T) {
	t.Parallel()

	clientRepo := newMockClientRepo()
	platformRepo := newMockPlatformRepo()

	clientID := uuid.New()
	platformID := uuid.New()
	clientRepo.clients[clientID.String()] = sqlc.Client{ID: clientID, Status: sqlc.ClientStatusSUSPENDED}
	platformRepo.platforms[platformID.String()] = sqlc.Platform{ID: platformID, Name: "HighLevel", Slug: "highlevel", Enabled: true}

	registry := providers.NewProviderRegistry()
	registry.Register(providers.NewHighLevelProvider("test-client", "test-secret", "https://example.com/callback", "", nil))

	svc := NewService(
		newMockIntegrationRepo(),
		newMockOAuthTokenRepo(),
		clientRepo,
		platformRepo,
		newMockOAuthStateRepo(),
		newMockPaymentProviderConfigRepo(),
		registry,
		"https://example.com/callback",
		ProviderConfigSettings{},
		zerolog.Nop(),
	)

	_, _, err := svc.BeginAuthorization(context.Background(), clientID, platformID)
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("status code = %s, want %s", status.Code(err), codes.FailedPrecondition)
	}
}

func TestHandleCallbackMissingCode(t *testing.T) {
	t.Parallel()

	svc := NewService(
		newMockIntegrationRepo(),
		newMockOAuthTokenRepo(),
		newMockClientRepo(),
		newMockPlatformRepo(),
		newMockOAuthStateRepo(),
		newMockPaymentProviderConfigRepo(),
		providers.NewProviderRegistry(),
		"https://example.com/callback",
		ProviderConfigSettings{},
		zerolog.Nop(),
	)

	_, err := svc.HandleCallback(context.Background(), "", "state")
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("status code = %s, want %s", status.Code(err), codes.InvalidArgument)
	}
}

func TestHandleCallbackMissingState(t *testing.T) {
	t.Parallel()

	svc := NewService(
		newMockIntegrationRepo(),
		newMockOAuthTokenRepo(),
		newMockClientRepo(),
		newMockPlatformRepo(),
		newMockOAuthStateRepo(),
		newMockPaymentProviderConfigRepo(),
		providers.NewProviderRegistry(),
		"https://example.com/callback",
		ProviderConfigSettings{},
		zerolog.Nop(),
	)

	_, err := svc.HandleCallback(context.Background(), "code", "")
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("status code = %s, want %s", status.Code(err), codes.InvalidArgument)
	}
}

func TestHandleCallbackInvalidState(t *testing.T) {
	t.Parallel()

	svc := NewService(
		newMockIntegrationRepo(),
		newMockOAuthTokenRepo(),
		newMockClientRepo(),
		newMockPlatformRepo(),
		newMockOAuthStateRepo(),
		newMockPaymentProviderConfigRepo(),
		providers.NewProviderRegistry(),
		"https://example.com/callback",
		ProviderConfigSettings{},
		zerolog.Nop(),
	)

	_, err := svc.HandleCallback(context.Background(), "code", "unknown-state")
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("status code = %s, want %s", status.Code(err), codes.InvalidArgument)
	}
}

func TestHandleCallbackExpiredState(t *testing.T) {
	t.Parallel()

	stateRepo := newMockOAuthStateRepo()
	clientID := uuid.New()
	platformID := uuid.New()
	stateRepo.states["expired-state"] = sqlc.OauthState{
		ID:         uuid.New(),
		State:      "expired-state",
		ClientID:   clientID,
		PlatformID: platformID,
		ExpiresAt:  time.Now().Add(-time.Minute),
	}

	svc := NewService(
		newMockIntegrationRepo(),
		newMockOAuthTokenRepo(),
		newMockClientRepo(),
		newMockPlatformRepo(),
		stateRepo,
		newMockPaymentProviderConfigRepo(),
		providers.NewProviderRegistry(),
		"https://example.com/callback",
		ProviderConfigSettings{},
		zerolog.Nop(),
	)

	_, err := svc.HandleCallback(context.Background(), "code", "expired-state")
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("status code = %s, want %s", status.Code(err), codes.InvalidArgument)
	}
}

func TestHandleCallbackConsumedState(t *testing.T) {
	t.Parallel()

	stateRepo := newMockOAuthStateRepo()
	clientID := uuid.New()
	platformID := uuid.New()
	stateRepo.states["consumed-state"] = sqlc.OauthState{
		ID:         uuid.New(),
		State:      "consumed-state",
		ClientID:   clientID,
		PlatformID: platformID,
		ExpiresAt:  time.Now().Add(time.Minute),
		ConsumedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	svc := NewService(
		newMockIntegrationRepo(),
		newMockOAuthTokenRepo(),
		newMockClientRepo(),
		newMockPlatformRepo(),
		stateRepo,
		newMockPaymentProviderConfigRepo(),
		providers.NewProviderRegistry(),
		"https://example.com/callback",
		ProviderConfigSettings{},
		zerolog.Nop(),
	)

	_, err := svc.HandleCallback(context.Background(), "code", "consumed-state")
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("status code = %s, want %s", status.Code(err), codes.InvalidArgument)
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
		newMockOAuthStateRepo(),
		newMockPaymentProviderConfigRepo(),
		registry,
		"https://example.com/callback",
		ProviderConfigSettings{},
		zerolog.Nop(),
	)

	_, err := svc.AuthorizationURL(context.Background(), uuid.New(), platformID, "test-state")
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("status code = %s, want %s", status.Code(err), codes.FailedPrecondition)
	}
}

// --- Stage 03: HighLevel Registration Lifecycle Tests ---

// newRegistrationTestService builds an OAuth service wired to in-memory mocks
// and a HighLevel provider whose OAuth and payment client point at the given
// mock server. It returns the service and the mock repos for assertions.
func newRegistrationTestService(t *testing.T, paymentHandler http.HandlerFunc) (*Service, *mockIntegrationRepo, *mockOAuthTokenRepo, *mockClientRepo, *mockPlatformRepo, *mockOAuthStateRepo, *mockPaymentProviderConfigRepo) {
	t.Helper()

	// Mock HighLevel server that handles both OAuth token exchange, user info,
	// and payment provider endpoints.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"access_token":"test-access-token",
				"refresh_token":"test-refresh-token",
				"expires_in":3600,
				"token_type":"Bearer",
				"scope":"read write",
				"locationId":"loc-123"
			}`))
		case "/v1/users/me":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"loc-123"}`))
		default:
			paymentHandler(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	paymentClient := providers.NewHighLevelPaymentProviderClient(srv.URL, nil)
	registry := providers.NewProviderRegistry()
	registry.Register(providers.NewHighLevelProviderWithURLs(
		"test-client",
		"test-secret",
		"https://example.com/callback",
		"",
		srv.URL+"/oauth/authorize",
		srv.URL+"/oauth/token",
		srv.URL+"/v1/users/me",
		paymentClient,
	))

	integrationRepo := newMockIntegrationRepo()
	tokenRepo := newMockOAuthTokenRepo()
	clientRepo := newMockClientRepo()
	platformRepo := newMockPlatformRepo()
	stateRepo := newMockOAuthStateRepo()
	configRepo := newMockPaymentProviderConfigRepo()

	svc := NewService(
		integrationRepo,
		tokenRepo,
		clientRepo,
		platformRepo,
		stateRepo,
		configRepo,
		registry,
		"https://example.com/callback",
		ProviderConfigSettings{
			Name:        "RVPay",
			Description: "RVPay payment provider",
			ImageURL:    "https://example.com/logo.jpg",
			PaymentsURL: "https://checkout.example.com/payment/checkout",
			QueryURL:    "https://api.example.com/payments/custom-provider/query",
		},
		zerolog.Nop(),
	)

	return svc, integrationRepo, tokenRepo, clientRepo, platformRepo, stateRepo, configRepo
}

// setupRegistrationContext populates the mocks with a valid client, platform,
// and OAuth state so ProcessCallback can be exercised.
func setupRegistrationContext(t *testing.T, clientRepo *mockClientRepo, platformRepo *mockPlatformRepo, stateRepo *mockOAuthStateRepo) (uuid.UUID, uuid.UUID, string) {
	t.Helper()

	clientID := uuid.New()
	platformID := uuid.New()
	clientRepo.clients[clientID.String()] = sqlc.Client{ID: clientID, Status: sqlc.ClientStatusACTIVE}
	platformRepo.platforms[platformID.String()] = sqlc.Platform{ID: platformID, Name: "HighLevel", Slug: "highlevel", Enabled: true}

	state := "test-state"
	_, err := stateRepo.Create(context.Background(), state, clientID, platformID, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("create state: %v", err)
	}

	return clientID, platformID, state
}

func TestProcessCallback_ProviderRegistrationSuccess(t *testing.T) {
	t.Parallel()

	// Mock HighLevel: both association and config creation succeed.
	svc, integrationRepo, tokenRepo, clientRepo, platformRepo, stateRepo, configRepo := newRegistrationTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/payments/custom-provider/provider":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true}`))
		case "/payments/custom-provider/connect":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"success":true}`))
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	clientID, platformID, state := setupRegistrationContext(t, clientRepo, platformRepo, stateRepo)

	result, err := svc.HandleCallback(context.Background(), "test-code", state)
	if err != nil {
		t.Fatalf("HandleCallback failed: %v", err)
	}

	if !result.ProviderRegistered {
		t.Fatal("ProviderRegistered should be true")
	}
	if result.ProviderRegistrationError != nil {
		t.Fatalf("ProviderRegistrationError should be nil, got %v", result.ProviderRegistrationError)
	}

	// Integration and token must be persisted.
	if len(integrationRepo.integrations) != 1 {
		t.Fatalf("expected 1 integration, got %d", len(integrationRepo.integrations))
	}
	if len(tokenRepo.tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokenRepo.tokens))
	}

	// Provider config must be persisted.
	if len(configRepo.configs) != 1 {
		t.Fatalf("expected 1 provider config, got %d", len(configRepo.configs))
	}
	for _, config := range configRepo.configs {
		if config.LocationID != "loc-123" {
			t.Errorf("LocationID = %q, want loc-123", config.LocationID)
		}
		if config.ProviderName != "RVPay" {
			t.Errorf("ProviderName = %q, want RVPay", config.ProviderName)
		}
		if config.QueryUrl != "https://api.example.com/payments/custom-provider/query" {
			t.Errorf("QueryUrl = %q, want configured URL", config.QueryUrl)
		}
		if config.PaymentsUrl != "https://checkout.example.com/payment/checkout" {
			t.Errorf("PaymentsUrl = %q, want configured URL", config.PaymentsUrl)
		}
		if config.SupportsSubscriptionSchedule {
			t.Error("SupportsSubscriptionSchedule should be false")
		}
		if config.ProviderApiKey == "" {
			t.Error("ProviderApiKey should not be empty")
		}
	}

	// The client and platform IDs must match.
	for _, integration := range integrationRepo.integrations {
		if integration.ClientID != clientID {
			t.Errorf("integration ClientID = %v, want %v", integration.ClientID, clientID)
		}
		if integration.PlatformID != platformID {
			t.Errorf("integration PlatformID = %v, want %v", integration.PlatformID, platformID)
		}
	}
}

func TestProcessCallback_ProviderAssociationFailure(t *testing.T) {
	t.Parallel()

	// Mock HighLevel: association fails with 500.
	svc, _, _, clientRepo, platformRepo, stateRepo, _ := newRegistrationTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/payments/custom-provider/provider":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"internal error"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	_, _, state := setupRegistrationContext(t, clientRepo, platformRepo, stateRepo)

	result, err := svc.HandleCallback(context.Background(), "test-code", state)
	if err != nil {
		t.Fatalf("HandleCallback failed: %v", err)
	}

	// OAuth installation succeeds even though registration fails.
	if result.ProviderRegistered {
		t.Fatal("ProviderRegistered should be false when association fails")
	}
	if result.ProviderRegistrationError == nil {
		t.Fatal("ProviderRegistrationError should be set when association fails")
	}
	if status.Code(result.ProviderRegistrationError) != codes.Internal {
		t.Fatalf("ProviderRegistrationError code = %s, want %s", status.Code(result.ProviderRegistrationError), codes.Internal)
	}
}

func TestProcessCallback_ProviderConfigFailure(t *testing.T) {
	t.Parallel()

	// Mock HighLevel: association succeeds, config creation fails with 500.
	svc, _, _, clientRepo, platformRepo, stateRepo, _ := newRegistrationTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/payments/custom-provider/provider":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true}`))
		case "/payments/custom-provider/connect":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"internal error"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	_, _, state := setupRegistrationContext(t, clientRepo, platformRepo, stateRepo)

	result, err := svc.HandleCallback(context.Background(), "test-code", state)
	if err != nil {
		t.Fatalf("HandleCallback failed: %v", err)
	}

	if result.ProviderRegistered {
		t.Fatal("ProviderRegistered should be false when config creation fails")
	}
	if result.ProviderRegistrationError == nil {
		t.Fatal("ProviderRegistrationError should be set when config creation fails")
	}
	if status.Code(result.ProviderRegistrationError) != codes.Internal {
		t.Fatalf("ProviderRegistrationError code = %s, want %s", status.Code(result.ProviderRegistrationError), codes.Internal)
	}
}

func TestProcessCallback_AlreadyConfiguredLocation(t *testing.T) {
	t.Parallel()

	// Mock HighLevel: association returns 400 (already exists), config
	// creation returns 422 (already exists), and fetch returns the existing
	// configuration.
	svc, _, _, clientRepo, platformRepo, stateRepo, configRepo := newRegistrationTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/payments/custom-provider/provider":
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"provider already exists"}`))
		case "/payments/custom-provider/connect":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write([]byte(`{"message":"config already exists"}`))
			} else if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{
					"name":"RVPay",
					"description":"RVPay payment provider",
					"imageUrl":"https://example.com/logo.jpg",
					"locationId":"loc-123",
					"queryUrl":"https://api.example.com/payments/custom-provider/query",
					"paymentsUrl":"https://checkout.example.com/payment/checkout",
					"supportsSubscriptionSchedule":false
				}`))
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	_, _, state := setupRegistrationContext(t, clientRepo, platformRepo, stateRepo)

	result, err := svc.HandleCallback(context.Background(), "test-code", state)
	if err != nil {
		t.Fatalf("HandleCallback failed: %v", err)
	}

	// The registration should succeed idempotently by fetching the existing
	// configuration.
	if !result.ProviderRegistered {
		t.Fatal("ProviderRegistered should be true for already-configured location")
	}
	if result.ProviderRegistrationError != nil {
		t.Fatalf("ProviderRegistrationError should be nil, got %v", result.ProviderRegistrationError)
	}

	// The config must be persisted locally.
	if len(configRepo.configs) != 1 {
		t.Fatalf("expected 1 provider config, got %d", len(configRepo.configs))
	}
	for _, config := range configRepo.configs {
		if config.LocationID != "loc-123" {
			t.Errorf("LocationID = %q, want loc-123", config.LocationID)
		}
	}
}

func TestRegisterProvider_RetrySafe(t *testing.T) {
	t.Parallel()

	// Mock HighLevel: first call to association fails with 500, second call
	// succeeds. This simulates a transient failure that can be retried.
	callCount := 0
	svc, integrationRepo, _, clientRepo, platformRepo, _, configRepo := newRegistrationTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/payments/custom-provider/provider":
			callCount++
			if callCount == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"message":"transient error"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true}`))
		case "/payments/custom-provider/connect":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	clientID := uuid.New()
	platformID := uuid.New()
	clientRepo.clients[clientID.String()] = sqlc.Client{ID: clientID, Status: sqlc.ClientStatusACTIVE}
	platformRepo.platforms[platformID.String()] = sqlc.Platform{ID: platformID, Name: "HighLevel", Slug: "highlevel", Enabled: true}

	// Create the integration directly (simulating a completed OAuth flow).
	integration, err := integrationRepo.Create(context.Background(), clientID, platformID, "loc-123", sqlc.IntegrationStatusACTIVE)
	if err != nil {
		t.Fatalf("create integration: %v", err)
	}

	// First attempt fails.
	err = svc.RegisterProvider(context.Background(), integration.ID, "loc-123", "test-access-token")
	if err == nil {
		t.Fatal("first RegisterProvider should fail")
	}
	if status.Code(err) != codes.Internal {
		t.Fatalf("first RegisterProvider error code = %s, want %s", status.Code(err), codes.Internal)
	}

	// No config should be persisted after failure.
	if len(configRepo.configs) != 0 {
		t.Fatalf("expected 0 configs after failure, got %d", len(configRepo.configs))
	}

	// Second attempt succeeds (retry-safe).
	err = svc.RegisterProvider(context.Background(), integration.ID, "loc-123", "test-access-token")
	if err != nil {
		t.Fatalf("second RegisterProvider failed: %v", err)
	}

	// Config should be persisted after success.
	if len(configRepo.configs) != 1 {
		t.Fatalf("expected 1 config after success, got %d", len(configRepo.configs))
	}
}

func TestRegisterProvider_MissingLocationID(t *testing.T) {
	t.Parallel()

	svc, _, _, _, _, _, _ := newRegistrationTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	err := svc.RegisterProvider(context.Background(), uuid.New(), "", "test-access-token")
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("error code = %s, want %s", status.Code(err), codes.InvalidArgument)
	}
}

func TestRegisterProvider_MissingAccessToken(t *testing.T) {
	t.Parallel()

	svc, _, _, _, _, _, _ := newRegistrationTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	err := svc.RegisterProvider(context.Background(), uuid.New(), "loc-123", "")
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("error code = %s, want %s", status.Code(err), codes.InvalidArgument)
	}
}

func TestRegisterProvider_IntegrationNotFound(t *testing.T) {
	t.Parallel()

	svc, _, _, _, _, _, _ := newRegistrationTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	err := svc.RegisterProvider(context.Background(), uuid.New(), "loc-123", "test-access-token")
	if status.Code(err) != codes.NotFound {
		t.Fatalf("error code = %s, want %s", status.Code(err), codes.NotFound)
	}
}

func TestRegisterProvider_NoConfigRepo(t *testing.T) {
	t.Parallel()

	// Build a service with a nil config repo.
	registry := providers.NewProviderRegistry()
	registry.Register(providers.NewHighLevelProvider("test-client", "test-secret", "https://example.com/callback", "", nil))

	svc := NewService(
		newMockIntegrationRepo(),
		newMockOAuthTokenRepo(),
		newMockClientRepo(),
		newMockPlatformRepo(),
		newMockOAuthStateRepo(),
		nil, // nil config repo
		registry,
		"https://example.com/callback",
		ProviderConfigSettings{},
		zerolog.Nop(),
	)

	err := svc.RegisterProvider(context.Background(), uuid.New(), "loc-123", "test-access-token")
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("error code = %s, want %s", status.Code(err), codes.FailedPrecondition)
	}
}
