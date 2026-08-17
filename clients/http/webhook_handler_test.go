package http

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/I-Frostbyte/rvpay-go/clients/providers"
	"github.com/I-Frostbyte/rvpay-go/clients/webhooks"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

// testWebhookEventRepo is an in-memory WebhookEventRepo for handler tests.
type testWebhookEventRepo struct {
	events map[string]sqlc.WebhookEvent
}

func newTestWebhookEventRepo() *testWebhookEventRepo {
	return &testWebhookEventRepo{events: make(map[string]sqlc.WebhookEvent)}
}

func (m *testWebhookEventRepo) Create(ctx context.Context, integrationID uuid.UUID, providerEventID, eventType string, payload []byte) (sqlc.WebhookEvent, error) {
	key := integrationID.String() + ":" + providerEventID
	if _, ok := m.events[key]; ok {
		return sqlc.WebhookEvent{}, repo.ErrDuplicate
	}
	event := sqlc.WebhookEvent{ID: uuid.New(), IntegrationID: integrationID, ProviderEventID: providerEventID, EventType: eventType, Payload: payload}
	m.events[key] = event
	return event, nil
}

func (m *testWebhookEventRepo) GetByIntegrationAndProvider(ctx context.Context, integrationID uuid.UUID, providerEventID string) (sqlc.WebhookEvent, error) {
	key := integrationID.String() + ":" + providerEventID
	event, ok := m.events[key]
	if !ok {
		return sqlc.WebhookEvent{}, repo.ErrNotFound
	}
	return event, nil
}

// testWebhookSubscriptionRepo is an in-memory WebhookSubscriptionRepo.
type testWebhookSubscriptionRepo struct {
	subscriptions map[string]sqlc.WebhookSubscription
}

func newTestWebhookSubscriptionRepo() *testWebhookSubscriptionRepo {
	return &testWebhookSubscriptionRepo{subscriptions: make(map[string]sqlc.WebhookSubscription)}
}

func (m *testWebhookSubscriptionRepo) Create(ctx context.Context, integrationID uuid.UUID, endpoint, secret string, status sqlc.WebhookSubscriptionStatus) (sqlc.WebhookSubscription, error) {
	sub := sqlc.WebhookSubscription{ID: uuid.New(), IntegrationID: integrationID, Endpoint: endpoint, Secret: secret, Status: status}
	m.subscriptions[sub.ID.String()] = sub
	return sub, nil
}
func (m *testWebhookSubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.WebhookSubscription, error) {
	sub, ok := m.subscriptions[id.String()]
	if !ok {
		return sqlc.WebhookSubscription{}, repo.ErrNotFound
	}
	return sub, nil
}
func (m *testWebhookSubscriptionRepo) GetByIntegrationIDAndEndpoint(ctx context.Context, integrationID uuid.UUID, endpoint string) (sqlc.WebhookSubscription, error) {
	for _, sub := range m.subscriptions {
		if sub.IntegrationID == integrationID {
			return sub, nil
		}
	}
	return sqlc.WebhookSubscription{}, repo.ErrNotFound
}
func (m *testWebhookSubscriptionRepo) ListByIntegrationID(ctx context.Context, integrationID uuid.UUID, limit, offset int32) ([]sqlc.WebhookSubscription, error) {
	return nil, nil
}
func (m *testWebhookSubscriptionRepo) ListActiveByIntegrationID(ctx context.Context, integrationID uuid.UUID) ([]sqlc.WebhookSubscription, error) {
	return nil, nil
}
func (m *testWebhookSubscriptionRepo) CountByIntegrationID(ctx context.Context, integrationID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *testWebhookSubscriptionRepo) Exists(ctx context.Context, integrationID uuid.UUID, endpoint string) (bool, error) {
	return false, nil
}
func (m *testWebhookSubscriptionRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.WebhookSubscriptionStatus) (sqlc.WebhookSubscription, error) {
	return sqlc.WebhookSubscription{}, nil
}
func (m *testWebhookSubscriptionRepo) UpdateLastDelivery(ctx context.Context, id uuid.UUID, lastDelivery pgtype.Timestamptz) (sqlc.WebhookSubscription, error) {
	return sqlc.WebhookSubscription{}, nil
}
func (m *testWebhookSubscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

// testWebhookIntegrationRepo is an in-memory IntegrationRepo for handler tests.
type testWebhookIntegrationRepo struct {
	integrations map[string]sqlc.Integration
}

func newTestWebhookIntegrationRepo() *testWebhookIntegrationRepo {
	return &testWebhookIntegrationRepo{integrations: make(map[string]sqlc.Integration)}
}

func (m *testWebhookIntegrationRepo) Create(ctx context.Context, clientID, platformID uuid.UUID, externalAccountID string, status sqlc.IntegrationStatus) (sqlc.Integration, error) {
	return sqlc.Integration{}, nil
}
func (m *testWebhookIntegrationRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Integration, error) {
	i, ok := m.integrations[id.String()]
	if !ok {
		return sqlc.Integration{}, repo.ErrNotFound
	}
	return i, nil
}
func (m *testWebhookIntegrationRepo) GetByClientAndPlatform(ctx context.Context, clientID, platformID uuid.UUID) (sqlc.Integration, error) {
	return sqlc.Integration{}, repo.ErrNotFound
}
func (m *testWebhookIntegrationRepo) GetByExternalAccountID(ctx context.Context, externalAccountID string) (sqlc.Integration, error) {
	return sqlc.Integration{}, repo.ErrNotFound
}
func (m *testWebhookIntegrationRepo) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	return nil, nil
}
func (m *testWebhookIntegrationRepo) ListByPlatform(ctx context.Context, platformID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	return nil, nil
}
func (m *testWebhookIntegrationRepo) ListActiveByClient(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	return nil, nil
}
func (m *testWebhookIntegrationRepo) CountByClient(ctx context.Context, clientID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *testWebhookIntegrationRepo) ExistsByClientAndPlatform(ctx context.Context, clientID, platformID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *testWebhookIntegrationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.IntegrationStatus) (sqlc.Integration, error) {
	return sqlc.Integration{}, nil
}
func (m *testWebhookIntegrationRepo) UpdateLastSyncAt(ctx context.Context, id uuid.UUID, lastSyncAt pgtype.Timestamptz) (sqlc.Integration, error) {
	return sqlc.Integration{}, nil
}
func (m *testWebhookIntegrationRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

// testPaymentProviderConfigRepo is an in-memory PaymentProviderConfigRepo for
// handler tests.
type testPaymentProviderConfigRepo struct {
	configs map[string]sqlc.PaymentProviderConfig
}

func newTestPaymentProviderConfigRepo() *testPaymentProviderConfigRepo {
	return &testPaymentProviderConfigRepo{configs: make(map[string]sqlc.PaymentProviderConfig)}
}

func (m *testPaymentProviderConfigRepo) Create(ctx context.Context, integrationID uuid.UUID, providerName, providerDescription, providerImageURL, locationID, queryURL, paymentsURL string, supportsSubscriptionSchedule bool, providerAPIKey string) (sqlc.PaymentProviderConfig, error) {
	return sqlc.PaymentProviderConfig{}, nil
}
func (m *testPaymentProviderConfigRepo) GetByIntegrationID(ctx context.Context, integrationID uuid.UUID) (sqlc.PaymentProviderConfig, error) {
	return sqlc.PaymentProviderConfig{}, repo.ErrNotFound
}
func (m *testPaymentProviderConfigRepo) GetByLocationID(ctx context.Context, locationID string) (sqlc.PaymentProviderConfig, error) {
	for _, c := range m.configs {
		if c.LocationID == locationID {
			return c, nil
		}
	}
	return sqlc.PaymentProviderConfig{}, repo.ErrNotFound
}
func (m *testPaymentProviderConfigRepo) GetByAPIKey(ctx context.Context, apiKey string) (sqlc.PaymentProviderConfig, error) {
	return sqlc.PaymentProviderConfig{}, repo.ErrNotFound
}
func (m *testPaymentProviderConfigRepo) Update(ctx context.Context, integrationID uuid.UUID, providerName, providerDescription, providerImageURL, locationID, queryURL, paymentsURL string, supportsSubscriptionSchedule bool, providerAPIKey string) (sqlc.PaymentProviderConfig, error) {
	return sqlc.PaymentProviderConfig{}, nil
}
func (m *testPaymentProviderConfigRepo) Delete(ctx context.Context, integrationID uuid.UUID) error {
	return nil
}

// testWebhookPlatformRepo is an in-memory PlatformRepo for handler tests.
type testWebhookPlatformRepo struct {
	platforms map[string]sqlc.Platform
}

func newTestWebhookPlatformRepo() *testWebhookPlatformRepo {
	return &testWebhookPlatformRepo{platforms: make(map[string]sqlc.Platform)}
}

func (m *testWebhookPlatformRepo) Create(ctx context.Context, name, displayName, slug string, enabled, oauthCapable, webhookCapable bool) (sqlc.Platform, error) {
	return sqlc.Platform{}, nil
}
func (m *testWebhookPlatformRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Platform, error) {
	p, ok := m.platforms[id.String()]
	if !ok {
		return sqlc.Platform{}, repo.ErrNotFound
	}
	return p, nil
}
func (m *testWebhookPlatformRepo) GetByName(ctx context.Context, name string) (sqlc.Platform, error) {
	return sqlc.Platform{}, repo.ErrNotFound
}
func (m *testWebhookPlatformRepo) GetBySlug(ctx context.Context, slug string) (sqlc.Platform, error) {
	return sqlc.Platform{}, repo.ErrNotFound
}
func (m *testWebhookPlatformRepo) List(ctx context.Context, limit, offset int32) ([]sqlc.Platform, error) {
	return nil, nil
}
func (m *testWebhookPlatformRepo) ListEnabled(ctx context.Context, limit, offset int32) ([]sqlc.Platform, error) {
	return nil, nil
}
func (m *testWebhookPlatformRepo) Count(ctx context.Context) (int64, error) { return 0, nil }
func (m *testWebhookPlatformRepo) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return false, nil
}
func (m *testWebhookPlatformRepo) Update(ctx context.Context, id uuid.UUID, name, displayName, slug string, enabled, oauthCapable, webhookCapable bool) (sqlc.Platform, error) {
	return sqlc.Platform{}, nil
}
func (m *testWebhookPlatformRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

// newTestWebhookHandler builds a WebhookHandler wired to in-memory mocks with a
// valid HighLevel provider registered using the given Ed25519 public key.
func newTestWebhookHandler(t *testing.T, publicKeyPEM string) (*WebhookHandler, *testWebhookEventRepo, *testWebhookSubscriptionRepo, *testPaymentProviderConfigRepo) {
	t.Helper()

	eventRepo := newTestWebhookEventRepo()
	subRepo := newTestWebhookSubscriptionRepo()
	configRepo := newTestPaymentProviderConfigRepo()

	registry := providers.NewProviderRegistry()
	registry.Register(providers.NewHighLevelProvider("test-client", "test-secret", "https://example.com/callback", publicKeyPEM, nil))

	svc := webhooks.NewService(
		newTestWebhookIntegrationRepo(),
		subRepo,
		eventRepo,
		newTestWebhookPlatformRepo(),
		configRepo,
		registry,
		zerolog.Nop(),
	)

	return NewWebhookHandler(svc, zerolog.Nop()), eventRepo, subRepo, configRepo
}

// testEd25519KeyPair generates a fresh Ed25519 key pair and returns the
// PEM-encoded public key and the raw private key for signing test payloads.
func testEd25519KeyPair(t *testing.T) (publicKeyPEM string, privateKey ed25519.PrivateKey) {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate ed25519 key: %v", err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}

	publicKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}))

	return publicKeyPEM, priv
}

// signBody signs the raw body with the private key and returns the base64
// encoded signature for the X-GHL-Signature header.
func signBody(t *testing.T, priv ed25519.PrivateKey, body []byte) string {
	t.Helper()
	sig := ed25519.Sign(priv, body)
	return base64.StdEncoding.EncodeToString(sig)
}

func TestWebhookHighLevel_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	publicKeyPEM, _ := testEd25519KeyPair(t)
	handler, _, _, _ := newTestWebhookHandler(t, publicKeyPEM)

	req := httptest.NewRequest(http.MethodGet, "/webhooks/highlevel", nil)
	rec := httptest.NewRecorder()

	handler.HighLevel(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestWebhookHighLevel_MissingSignature(t *testing.T) {
	t.Parallel()

	publicKeyPEM, _ := testEd25519KeyPair(t)
	handler, _, _, _ := newTestWebhookHandler(t, publicKeyPEM)

	body := `{"eventId":"evt_123","eventType":"integration.installed","integrationId":"00000000-0000-0000-0000-000000000001"}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks/highlevel", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.HighLevel(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestWebhookHighLevel_InvalidSignature(t *testing.T) {
	t.Parallel()

	publicKeyPEM, _ := testEd25519KeyPair(t)
	handler, _, _, _ := newTestWebhookHandler(t, publicKeyPEM)

	body := `{"eventId":"evt_123","eventType":"integration.installed","integrationId":"00000000-0000-0000-0000-000000000001"}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks/highlevel", strings.NewReader(body))
	req.Header.Set("X-GHL-Signature", "aW52YWxpZA==")
	rec := httptest.NewRecorder()

	handler.HighLevel(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestWebhookHighLevel_ValidSignedRequest(t *testing.T) {
	t.Parallel()

	publicKeyPEM, priv := testEd25519KeyPair(t)
	handler, _, subRepo, configRepo := newTestWebhookHandler(t, publicKeyPEM)

	integrationID := uuid.New()
	subRepo.subscriptions[integrationID.String()] = sqlc.WebhookSubscription{
		ID:            uuid.New(),
		IntegrationID: integrationID,
		Endpoint:      "/webhooks/highlevel",
		Status:        sqlc.WebhookSubscriptionStatusACTIVE,
	}
	configRepo.configs["kSRxQkM72aCeYz19uw79"] = sqlc.PaymentProviderConfig{
		IntegrationID: integrationID,
		LocationID:    "kSRxQkM72aCeYz19uw79",
	}

	body := `{"type":"INSTALL","appId":"6a5f8aafdb5067f4319b1bb4","versionId":"6a5f8aafdb5067f4319b1bb4","installType":"Location","locationId":"kSRxQkM72aCeYz19uw79","companyId":"f3rBPevH93JANjvqtrK0","userId":"K6MePugfKQPdgEicKzVJ","companyName":"evaristustambua@gmail.com","isWhitelabelCompany":false,"whitelabelDetails":{"logoUrl":"","domain":""},"trial":{},"timestamp":"2026-08-17T09:06:59.366Z","webhookId":"f4ef22e3-c4c1-4ce5-996d-297890460e7d"}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks/highlevel", strings.NewReader(body))
	req.Header.Set("X-GHL-Signature", signBody(t, priv, []byte(body)))
	rec := httptest.NewRecorder()

	handler.HighLevel(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestWebhookHighLevel_DuplicateEvent(t *testing.T) {
	t.Parallel()

	publicKeyPEM, priv := testEd25519KeyPair(t)
	handler, eventRepo, subRepo, configRepo := newTestWebhookHandler(t, publicKeyPEM)

	integrationID := uuid.New()
	subRepo.subscriptions[integrationID.String()] = sqlc.WebhookSubscription{
		ID:            uuid.New(),
		IntegrationID: integrationID,
		Endpoint:      "/webhooks/highlevel",
		Status:        sqlc.WebhookSubscriptionStatusACTIVE,
	}
	configRepo.configs["kSRxQkM72aCeYz19uw79"] = sqlc.PaymentProviderConfig{
		IntegrationID: integrationID,
		LocationID:    "kSRxQkM72aCeYz19uw79",
	}

	body := `{"type":"INSTALL","appId":"6a5f8aafdb5067f4319b1bb4","versionId":"6a5f8aafdb5067f4319b1bb4","installType":"Location","locationId":"kSRxQkM72aCeYz19uw79","companyId":"f3rBPevH93JANjvqtrK0","userId":"K6MePugfKQPdgEicKzVJ","companyName":"evaristustambua@gmail.com","isWhitelabelCompany":false,"whitelabelDetails":{"logoUrl":"","domain":""},"trial":{},"timestamp":"2026-08-17T09:06:59.366Z","webhookId":"f4ef22e3-c4c1-4ce5-996d-297890460e7d"}`
	sig := signBody(t, priv, []byte(body))

	// First delivery succeeds.
	req1 := httptest.NewRequest(http.MethodPost, "/webhooks/highlevel", strings.NewReader(body))
	req1.Header.Set("X-GHL-Signature", sig)
	rec1 := httptest.NewRecorder()
	handler.HighLevel(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first delivery status = %d, want %d", rec1.Code, http.StatusOK)
	}

	// Second (duplicate) delivery is acknowledged as success (idempotent).
	req2 := httptest.NewRequest(http.MethodPost, "/webhooks/highlevel", strings.NewReader(body))
	req2.Header.Set("X-GHL-Signature", sig)
	rec2 := httptest.NewRecorder()
	handler.HighLevel(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("duplicate delivery status = %d, want %d", rec2.Code, http.StatusOK)
	}

	// Only one event should have been recorded.
	if len(eventRepo.events) != 1 {
		t.Fatalf("recorded events = %d, want 1", len(eventRepo.events))
	}
}

func TestWebhookHighLevel_MalformedJSON(t *testing.T) {
	t.Parallel()

	publicKeyPEM, priv := testEd25519KeyPair(t)
	handler, _, _, _ := newTestWebhookHandler(t, publicKeyPEM)

	body := `{invalid json`
	req := httptest.NewRequest(http.MethodPost, "/webhooks/highlevel", strings.NewReader(body))
	req.Header.Set("X-GHL-Signature", signBody(t, priv, []byte(body)))
	rec := httptest.NewRecorder()

	handler.HighLevel(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
