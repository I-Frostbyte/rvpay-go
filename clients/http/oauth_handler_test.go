package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/I-Frostbyte/rvpay-go/clients/oauth"
	"github.com/I-Frostbyte/rvpay-go/clients/providers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

// testOAuthStateRepo is an in-memory OAuthStateRepo for handler tests.
type testOAuthStateRepo struct {
	states map[string]sqlc.OauthState
}

func newTestOAuthStateRepo() *testOAuthStateRepo {
	return &testOAuthStateRepo{states: make(map[string]sqlc.OauthState)}
}

func (m *testOAuthStateRepo) Create(ctx context.Context, state string, clientID, platformID uuid.UUID, expiresAt time.Time) (sqlc.OauthState, error) {
	record := sqlc.OauthState{ID: uuid.New(), State: state, ClientID: clientID, PlatformID: platformID, ExpiresAt: expiresAt}
	m.states[state] = record
	return record, nil
}

func (m *testOAuthStateRepo) GetByState(ctx context.Context, state string) (sqlc.OauthState, error) {
	record, ok := m.states[state]
	if !ok {
		return sqlc.OauthState{}, repo.ErrNotFound
	}
	return record, nil
}

func (m *testOAuthStateRepo) Consume(ctx context.Context, state string) (sqlc.OauthState, error) {
	record, ok := m.states[state]
	if !ok {
		return sqlc.OauthState{}, repo.ErrNotFound
	}
	if record.ConsumedAt.Valid || time.Now().After(record.ExpiresAt) {
		return sqlc.OauthState{}, repo.ErrNotFound
	}
	record.ConsumedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	m.states[state] = record
	return record, nil
}

func (m *testOAuthStateRepo) DeleteExpired(ctx context.Context) (int64, error) { return 0, nil }

// testOAuthClientRepo is an in-memory ClientRepo for handler tests.
type testOAuthClientRepo struct {
	clients map[string]sqlc.Client
}

func newTestOAuthClientRepo() *testOAuthClientRepo {
	return &testOAuthClientRepo{clients: make(map[string]sqlc.Client)}
}

func (m *testOAuthClientRepo) Create(ctx context.Context, name string, status sqlc.ClientStatus) (sqlc.Client, error) {
	return sqlc.Client{}, nil
}
func (m *testOAuthClientRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Client, error) {
	c, ok := m.clients[id.String()]
	if !ok {
		return sqlc.Client{}, repo.ErrNotFound
	}
	return c, nil
}
func (m *testOAuthClientRepo) GetByName(ctx context.Context, name string) (sqlc.Client, error) {
	return sqlc.Client{}, repo.ErrNotFound
}
func (m *testOAuthClientRepo) List(ctx context.Context, limit, offset int32) ([]sqlc.Client, error) {
	return nil, nil
}
func (m *testOAuthClientRepo) ListActive(ctx context.Context, limit, offset int32) ([]sqlc.Client, error) {
	return nil, nil
}
func (m *testOAuthClientRepo) Count(ctx context.Context) (int64, error) { return 0, nil }
func (m *testOAuthClientRepo) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	_, ok := m.clients[id.String()]
	return ok, nil
}
func (m *testOAuthClientRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.ClientStatus) (sqlc.Client, error) {
	return sqlc.Client{}, nil
}
func (m *testOAuthClientRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

// testOAuthPlatformRepo is an in-memory PlatformRepo for handler tests.
type testOAuthPlatformRepo struct {
	platforms map[string]sqlc.Platform
}

func newTestOAuthPlatformRepo() *testOAuthPlatformRepo {
	return &testOAuthPlatformRepo{platforms: make(map[string]sqlc.Platform)}
}

func (m *testOAuthPlatformRepo) Create(ctx context.Context, name, displayName, slug string, enabled, oauthCapable, webhookCapable bool) (sqlc.Platform, error) {
	return sqlc.Platform{}, nil
}
func (m *testOAuthPlatformRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Platform, error) {
	p, ok := m.platforms[id.String()]
	if !ok {
		return sqlc.Platform{}, repo.ErrNotFound
	}
	return p, nil
}
func (m *testOAuthPlatformRepo) GetByName(ctx context.Context, name string) (sqlc.Platform, error) {
	return sqlc.Platform{}, repo.ErrNotFound
}
func (m *testOAuthPlatformRepo) GetBySlug(ctx context.Context, slug string) (sqlc.Platform, error) {
	return sqlc.Platform{}, repo.ErrNotFound
}
func (m *testOAuthPlatformRepo) List(ctx context.Context, limit, offset int32) ([]sqlc.Platform, error) {
	return nil, nil
}
func (m *testOAuthPlatformRepo) ListEnabled(ctx context.Context, limit, offset int32) ([]sqlc.Platform, error) {
	return nil, nil
}
func (m *testOAuthPlatformRepo) Count(ctx context.Context) (int64, error) { return 0, nil }
func (m *testOAuthPlatformRepo) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return false, nil
}
func (m *testOAuthPlatformRepo) Update(ctx context.Context, id uuid.UUID, name, displayName, slug string, enabled, oauthCapable, webhookCapable bool) (sqlc.Platform, error) {
	return sqlc.Platform{}, nil
}
func (m *testOAuthPlatformRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

// testOAuthIntegrationRepo is an in-memory IntegrationRepo for handler tests.
type testOAuthIntegrationRepo struct {
	integrations map[string]sqlc.Integration
}

func newTestOAuthIntegrationRepo() *testOAuthIntegrationRepo {
	return &testOAuthIntegrationRepo{integrations: make(map[string]sqlc.Integration)}
}

func (m *testOAuthIntegrationRepo) Create(ctx context.Context, clientID, platformID uuid.UUID, externalAccountID string, status sqlc.IntegrationStatus) (sqlc.Integration, error) {
	integration := sqlc.Integration{ID: uuid.New(), ClientID: clientID, PlatformID: platformID, ExternalAccountID: externalAccountID, Status: status, InstalledAt: time.Now()}
	m.integrations[integration.ID.String()] = integration
	return integration, nil
}
func (m *testOAuthIntegrationRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Integration, error) {
	i, ok := m.integrations[id.String()]
	if !ok {
		return sqlc.Integration{}, repo.ErrNotFound
	}
	return i, nil
}
func (m *testOAuthIntegrationRepo) GetByClientAndPlatform(ctx context.Context, clientID, platformID uuid.UUID) (sqlc.Integration, error) {
	for _, i := range m.integrations {
		if i.ClientID == clientID && i.PlatformID == platformID {
			return i, nil
		}
	}
	return sqlc.Integration{}, repo.ErrNotFound
}
func (m *testOAuthIntegrationRepo) GetByExternalAccountID(ctx context.Context, externalAccountID string) (sqlc.Integration, error) {
	return sqlc.Integration{}, repo.ErrNotFound
}
func (m *testOAuthIntegrationRepo) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	return nil, nil
}
func (m *testOAuthIntegrationRepo) ListByPlatform(ctx context.Context, platformID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	return nil, nil
}
func (m *testOAuthIntegrationRepo) ListActiveByClient(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	return nil, nil
}
func (m *testOAuthIntegrationRepo) CountByClient(ctx context.Context, clientID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *testOAuthIntegrationRepo) ExistsByClientAndPlatform(ctx context.Context, clientID, platformID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *testOAuthIntegrationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.IntegrationStatus) (sqlc.Integration, error) {
	return sqlc.Integration{}, nil
}
func (m *testOAuthIntegrationRepo) UpdateLastSyncAt(ctx context.Context, id uuid.UUID, lastSyncAt pgtype.Timestamptz) (sqlc.Integration, error) {
	return sqlc.Integration{}, nil
}
func (m *testOAuthIntegrationRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

// testOAuthTokenRepo is an in-memory OAuthTokenRepo for handler tests.
type testOAuthTokenRepo struct {
	tokens map[string]sqlc.OauthToken
}

func newTestOAuthTokenRepo() *testOAuthTokenRepo {
	return &testOAuthTokenRepo{tokens: make(map[string]sqlc.OauthToken)}
}

func (m *testOAuthTokenRepo) Create(ctx context.Context, integrationID uuid.UUID, accessToken, refreshToken string, expiresAt time.Time, scope, tokenType string) (sqlc.OauthToken, error) {
	token := sqlc.OauthToken{ID: uuid.New(), IntegrationID: integrationID, AccessToken: accessToken, RefreshToken: refreshToken, ExpiresAt: expiresAt, Scope: scope, TokenType: tokenType}
	m.tokens[token.ID.String()] = token
	return token, nil
}
func (m *testOAuthTokenRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.OauthToken, error) {
	return sqlc.OauthToken{}, repo.ErrNotFound
}
func (m *testOAuthTokenRepo) GetByIntegrationID(ctx context.Context, integrationID uuid.UUID) (sqlc.OauthToken, error) {
	return sqlc.OauthToken{}, repo.ErrNotFound
}
func (m *testOAuthTokenRepo) ExistsByIntegrationID(ctx context.Context, integrationID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *testOAuthTokenRepo) Update(ctx context.Context, id uuid.UUID, accessToken, refreshToken string, expiresAt time.Time, scope, tokenType string) (sqlc.OauthToken, error) {
	return sqlc.OauthToken{}, nil
}
func (m *testOAuthTokenRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *testOAuthTokenRepo) DeleteByIntegrationID(ctx context.Context, integrationID uuid.UUID) error {
	return nil
}

// newTestOAuthHandler builds an OAuthHandler wired to in-memory mocks with a
// valid HighLevel provider registered.
func newTestOAuthHandler(t *testing.T) (*OAuthHandler, *testOAuthStateRepo, *testOAuthClientRepo, *testOAuthPlatformRepo) {
	t.Helper()

	clientRepo := newTestOAuthClientRepo()
	platformRepo := newTestOAuthPlatformRepo()
	stateRepo := newTestOAuthStateRepo()

	registry := providers.NewProviderRegistry()
	registry.Register(providers.NewHighLevelProvider("test-client", "test-secret", "https://example.com/callback", "", nil))

	svc := oauth.NewService(
		newTestOAuthIntegrationRepo(),
		newTestOAuthTokenRepo(),
		clientRepo,
		platformRepo,
		stateRepo,
		nil,
		registry,
		"https://example.com/callback",
		oauth.ProviderConfigSettings{},
		zerolog.Nop(),
	)

	return NewOAuthHandler(svc, zerolog.Nop()), stateRepo, clientRepo, platformRepo
}

func TestOAuthCallback_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler, _, _, _ := newTestOAuthHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/oauth/callback?code=abc&state=xyz", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestOAuthCallback_MissingCode(t *testing.T) {
	t.Parallel()

	handler, _, _, _ := newTestOAuthHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?state=xyz", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestOAuthCallback_MissingState(t *testing.T) {
	t.Parallel()

	handler, _, _, _ := newTestOAuthHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?code=abc", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestOAuthCallback_InvalidState(t *testing.T) {
	t.Parallel()

	handler, _, _, _ := newTestOAuthHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?code=abc&state=unknown-state", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestOAuthCallback_ValidStateButClientNotFound(t *testing.T) {
	t.Parallel()

	handler, stateRepo, _, platformRepo := newTestOAuthHandler(t)

	clientID := uuid.New()
	platformID := uuid.New()
	platformRepo.platforms[platformID.String()] = sqlc.Platform{ID: platformID, Name: "HighLevel", Slug: "highlevel", Enabled: true}

	// Create a valid state but no client exists.
	_, err := stateRepo.Create(context.Background(), "valid-state", clientID, platformID, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("create state: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?code=abc&state=valid-state", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestOAuthCallback_ExpiredState(t *testing.T) {
	t.Parallel()

	handler, stateRepo, clientRepo, platformRepo := newTestOAuthHandler(t)

	clientID := uuid.New()
	platformID := uuid.New()
	clientRepo.clients[clientID.String()] = sqlc.Client{ID: clientID, Status: sqlc.ClientStatusACTIVE}
	platformRepo.platforms[platformID.String()] = sqlc.Platform{ID: platformID, Name: "HighLevel", Slug: "highlevel", Enabled: true}

	// Create an already-expired state.
	_, err := stateRepo.Create(context.Background(), "expired-state", clientID, platformID, time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatalf("create state: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?code=abc&state=expired-state", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
