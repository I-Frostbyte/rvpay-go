package providers

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

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

func TestVerifyRequest_ValidSignature(t *testing.T) {
	t.Parallel()

	publicKeyPEM, priv := testEd25519KeyPair(t)
	provider := NewHighLevelWebhookProvider(publicKeyPEM)

	body := []byte(`{"eventId":"evt_123","eventType":"integration.installed","integrationId":"` + "00000000-0000-0000-0000-000000000001" + `"}`)
	headers := map[string]string{
		"X-GHL-Signature": signBody(t, priv, body),
	}

	if err := provider.VerifyRequest(context.Background(), headers, body); err != nil {
		t.Fatalf("VerifyRequest failed for valid signature: %v", err)
	}
}

func TestVerifyRequest_InvalidSignature(t *testing.T) {
	t.Parallel()

	publicKeyPEM, priv := testEd25519KeyPair(t)
	provider := NewHighLevelWebhookProvider(publicKeyPEM)

	body := []byte(`{"eventId":"evt_123"}`)
	// Sign a different body so the signature does not match.
	otherBody := []byte(`{"eventId":"evt_456"}`)
	headers := map[string]string{
		"X-GHL-Signature": signBody(t, priv, otherBody),
	}

	if err := provider.VerifyRequest(context.Background(), headers, body); err == nil {
		t.Fatal("VerifyRequest should reject an invalid signature")
	}
}

func TestVerifyRequest_ModifiedBody(t *testing.T) {
	t.Parallel()

	publicKeyPEM, priv := testEd25519KeyPair(t)
	provider := NewHighLevelWebhookProvider(publicKeyPEM)

	body := []byte(`{"eventId":"evt_123","amount":100}`)
	headers := map[string]string{
		"X-GHL-Signature": signBody(t, priv, body),
	}

	// Tamper with the body after signing.
	modifiedBody := []byte(`{"eventId":"evt_123","amount":999}`)
	if err := provider.VerifyRequest(context.Background(), headers, modifiedBody); err == nil {
		t.Fatal("VerifyRequest should reject a modified body")
	}
}

func TestVerifyRequest_MissingSignature(t *testing.T) {
	t.Parallel()

	publicKeyPEM, _ := testEd25519KeyPair(t)
	provider := NewHighLevelWebhookProvider(publicKeyPEM)

	body := []byte(`{"eventId":"evt_123"}`)
	headers := map[string]string{}

	if err := provider.VerifyRequest(context.Background(), headers, body); err == nil {
		t.Fatal("VerifyRequest should reject a missing signature")
	}
}

func TestVerifyRequest_MalformedSignature(t *testing.T) {
	t.Parallel()

	publicKeyPEM, _ := testEd25519KeyPair(t)
	provider := NewHighLevelWebhookProvider(publicKeyPEM)

	body := []byte(`{"eventId":"evt_123"}`)
	headers := map[string]string{
		"X-GHL-Signature": "not-base64!!!",
	}

	if err := provider.VerifyRequest(context.Background(), headers, body); err == nil {
		t.Fatal("VerifyRequest should reject a malformed signature")
	}
}

func TestVerifyRequest_WrongPublicKey(t *testing.T) {
	t.Parallel()

	// Sign with one key, verify with a different key.
	_, priv := testEd25519KeyPair(t)
	otherPublicKeyPEM, _ := testEd25519KeyPair(t)
	provider := NewHighLevelWebhookProvider(otherPublicKeyPEM)

	body := []byte(`{"eventId":"evt_123"}`)
	headers := map[string]string{
		"X-GHL-Signature": signBody(t, priv, body),
	}

	if err := provider.VerifyRequest(context.Background(), headers, body); err == nil {
		t.Fatal("VerifyRequest should reject a signature from a different key")
	}
}

func TestVerifyRequest_EmptyBody(t *testing.T) {
	t.Parallel()

	publicKeyPEM, priv := testEd25519KeyPair(t)
	provider := NewHighLevelWebhookProvider(publicKeyPEM)

	body := []byte{}
	headers := map[string]string{
		"X-GHL-Signature": signBody(t, priv, body),
	}

	if err := provider.VerifyRequest(context.Background(), headers, body); err != nil {
		t.Fatalf("VerifyRequest failed for empty body with valid signature: %v", err)
	}
}

func TestVerifyRequest_FormattingPreserved(t *testing.T) {
	t.Parallel()

	publicKeyPEM, priv := testEd25519KeyPair(t)
	provider := NewHighLevelWebhookProvider(publicKeyPEM)

	// The signature is over the exact raw bytes, including whitespace and
	// key ordering. Re-marshaling would change the bytes and invalidate it.
	body := []byte("{\n  \"eventId\": \"evt_123\",\n  \"eventType\": \"integration.installed\"\n}")
	headers := map[string]string{
		"X-GHL-Signature": signBody(t, priv, body),
	}

	if err := provider.VerifyRequest(context.Background(), headers, body); err != nil {
		t.Fatalf("VerifyRequest failed for raw body with preserved formatting: %v", err)
	}
}

func TestVerifyRequest_NoPublicKeyConfigured(t *testing.T) {
	t.Parallel()

	provider := NewHighLevelWebhookProvider("")
	body := []byte(`{"eventId":"evt_123"}`)
	headers := map[string]string{
		"X-GHL-Signature": "c2lnbmF0dXJl",
	}

	if err := provider.VerifyRequest(context.Background(), headers, body); err == nil {
		t.Fatal("VerifyRequest should fail when no public key is configured")
	}
}

func TestParseEvent(t *testing.T) {
	t.Parallel()

	provider := NewHighLevelWebhookProvider("")
	// Exact GHL INSTALL payload.
	body := []byte(`{"type":"INSTALL","appId":"6a5f8aafdb5067f4319b1bb4","versionId":"6a5f8aafdb5067f4319b1bb4","installType":"Location","locationId":"kSRxQkM72aCeYz19uw79","companyId":"f3rBPevH93JANjvqtrK0","userId":"K6MePugfKQPdgEicKzVJ","companyName":"evaristustambua@gmail.com","isWhitelabelCompany":false,"whitelabelDetails":{"logoUrl":"","domain":""},"trial":{},"timestamp":"2026-08-17T09:06:59.366Z","webhookId":"f4ef22e3-c4c1-4ce5-996d-297890460e7d"}`)

	event, err := provider.ParseEvent(context.Background(), body)
	if err != nil {
		t.Fatalf("ParseEvent failed: %v", err)
	}

	if event.Provider != "highlevel" {
		t.Fatalf("Provider = %s, want highlevel", event.Provider)
	}
	if event.EventType != "INSTALL" {
		t.Fatalf("EventType = %s, want INSTALL", event.EventType)
	}
	if event.ProviderEventID != "f4ef22e3-c4c1-4ce5-996d-297890460e7d" {
		t.Fatalf("ProviderEventID = %s, want webhookId", event.ProviderEventID)
	}
	if event.IntegrationID != "6a5f8aafdb5067f4319b1bb4" {
		t.Fatalf("IntegrationID = %s, want appId", event.IntegrationID)
	}
	if event.ClientID != "f3rBPevH93JANjvqtrK0" {
		t.Fatalf("ClientID = %s, want companyId", event.ClientID)
	}
	if event.LocationID != "kSRxQkM72aCeYz19uw79" {
		t.Fatalf("LocationID = %s, want locationId", event.LocationID)
	}
	// 2026-08-17T09:06:59.366Z in Unix seconds.
	if event.ReceivedAt != 1786957619 {
		t.Fatalf("ReceivedAt = %d, want 1786957619", event.ReceivedAt)
	}
}

func TestParseEvent_MalformedJSON(t *testing.T) {
	t.Parallel()

	provider := NewHighLevelWebhookProvider("")
	body := []byte(`{invalid json`)

	if _, err := provider.ParseEvent(context.Background(), body); err == nil {
		t.Fatal("ParseEvent should reject malformed JSON")
	}
}

// --- INSTALL webhook dispatcher tests ---

// testDispatcherIntegrationRepo is an in-memory IntegrationRepo for dispatcher tests.
type testDispatcherIntegrationRepo struct {
	integrations map[string]sqlc.Integration
}

func newTestDispatcherIntegrationRepo() *testDispatcherIntegrationRepo {
	return &testDispatcherIntegrationRepo{integrations: make(map[string]sqlc.Integration)}
}

func (m *testDispatcherIntegrationRepo) Create(ctx context.Context, clientID, platformID uuid.UUID, externalAccountID string, status sqlc.IntegrationStatus) (sqlc.Integration, error) {
	return sqlc.Integration{}, nil
}
func (m *testDispatcherIntegrationRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Integration, error) {
	i, ok := m.integrations[id.String()]
	if !ok {
		return sqlc.Integration{}, repo.ErrNotFound
	}
	return i, nil
}
func (m *testDispatcherIntegrationRepo) GetByClientAndPlatform(ctx context.Context, clientID, platformID uuid.UUID) (sqlc.Integration, error) {
	return sqlc.Integration{}, repo.ErrNotFound
}
func (m *testDispatcherIntegrationRepo) GetByExternalAccountID(ctx context.Context, externalAccountID string) (sqlc.Integration, error) {
	for _, i := range m.integrations {
		if i.ExternalAccountID == externalAccountID {
			return i, nil
		}
	}
	return sqlc.Integration{}, repo.ErrNotFound
}
func (m *testDispatcherIntegrationRepo) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	return nil, nil
}
func (m *testDispatcherIntegrationRepo) ListByPlatform(ctx context.Context, platformID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	return nil, nil
}
func (m *testDispatcherIntegrationRepo) ListActiveByClient(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	return nil, nil
}
func (m *testDispatcherIntegrationRepo) CountByClient(ctx context.Context, clientID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *testDispatcherIntegrationRepo) ExistsByClientAndPlatform(ctx context.Context, clientID, platformID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *testDispatcherIntegrationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.IntegrationStatus) (sqlc.Integration, error) {
	return sqlc.Integration{}, nil
}
func (m *testDispatcherIntegrationRepo) UpdateLastSyncAt(ctx context.Context, id uuid.UUID, lastSyncAt pgtype.Timestamptz) (sqlc.Integration, error) {
	return sqlc.Integration{}, nil
}
func (m *testDispatcherIntegrationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// testDispatcherConfigRepo is an in-memory PaymentProviderConfigRepo for dispatcher tests.
type testDispatcherConfigRepo struct {
	configs map[string]sqlc.PaymentProviderConfig
}

func newTestDispatcherConfigRepo() *testDispatcherConfigRepo {
	return &testDispatcherConfigRepo{configs: make(map[string]sqlc.PaymentProviderConfig)}
}

func (m *testDispatcherConfigRepo) Create(ctx context.Context, integrationID uuid.UUID, providerName, providerDescription, providerImageURL, locationID, queryURL, paymentsURL string, supportsSubscriptionSchedule bool, providerAPIKey string) (sqlc.PaymentProviderConfig, error) {
	// Idempotent: if a config already exists for this integration, return ErrDuplicate.
	for _, c := range m.configs {
		if c.IntegrationID == integrationID {
			return sqlc.PaymentProviderConfig{}, repo.ErrDuplicate
		}
	}
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
func (m *testDispatcherConfigRepo) GetByIntegrationID(ctx context.Context, integrationID uuid.UUID) (sqlc.PaymentProviderConfig, error) {
	for _, c := range m.configs {
		if c.IntegrationID == integrationID {
			return c, nil
		}
	}
	return sqlc.PaymentProviderConfig{}, repo.ErrNotFound
}
func (m *testDispatcherConfigRepo) GetByLocationID(ctx context.Context, locationID string) (sqlc.PaymentProviderConfig, error) {
	for _, c := range m.configs {
		if c.LocationID == locationID {
			return c, nil
		}
	}
	return sqlc.PaymentProviderConfig{}, repo.ErrNotFound
}
func (m *testDispatcherConfigRepo) GetByAPIKey(ctx context.Context, apiKey string) (sqlc.PaymentProviderConfig, error) {
	return sqlc.PaymentProviderConfig{}, repo.ErrNotFound
}
func (m *testDispatcherConfigRepo) Update(ctx context.Context, integrationID uuid.UUID, providerName, providerDescription, providerImageURL, locationID, queryURL, paymentsURL string, supportsSubscriptionSchedule bool, providerAPIKey string) (sqlc.PaymentProviderConfig, error) {
	return sqlc.PaymentProviderConfig{}, nil
}
func (m *testDispatcherConfigRepo) Delete(ctx context.Context, integrationID uuid.UUID) error {
	return nil
}

// testDispatcherLogger is a no-op Logger for dispatcher tests.
type testDispatcherLogger struct{}

func (l *testDispatcherLogger) Info(msg string, args ...interface{}) {}

func TestDispatch_InstallCreatesConfig(t *testing.T) {
	t.Parallel()

	integrationRepo := newTestDispatcherIntegrationRepo()
	configRepo := newTestDispatcherConfigRepo()

	clientID := uuid.New()
	platformID := uuid.New()
	integrationID := uuid.New()
	// The integration has external_account_id = GHL locationId (activated).
	integrationRepo.integrations[integrationID.String()] = sqlc.Integration{
		ID:                integrationID,
		ClientID:          clientID,
		PlatformID:        platformID,
		ExternalAccountID: "loc-123",
		Status:            sqlc.IntegrationStatusACTIVE,
	}

	dispatcher := NewHighLevelWebhookDispatcher(
		&testDispatcherLogger{},
		integrationRepo,
		configRepo,
		ProviderConfigSettings{Name: "RVPay", QueryURL: "https://api.example.com/query", PaymentsURL: "https://checkout.example.com/pay"},
	)

	event := &WebhookEvent{
		Provider:        "highlevel",
		EventType:       "INSTALL",
		ProviderEventID: "evt-1",
		LocationID:      "loc-123",
	}

	err := dispatcher.Dispatch(context.Background(), event)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	// The config must be created for the resolved integration.
	if len(configRepo.configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configRepo.configs))
	}
	for _, c := range configRepo.configs {
		if c.IntegrationID != integrationID {
			t.Fatalf("config IntegrationID = %v, want %v", c.IntegrationID, integrationID)
		}
		if c.LocationID != "loc-123" {
			t.Fatalf("config LocationID = %q, want loc-123", c.LocationID)
		}
		if c.ProviderName != "RVPay" {
			t.Fatalf("config ProviderName = %q, want RVPay", c.ProviderName)
		}
	}
}

func TestDispatch_InstallReusesExistingConfig(t *testing.T) {
	t.Parallel()

	integrationRepo := newTestDispatcherIntegrationRepo()
	configRepo := newTestDispatcherConfigRepo()

	clientID := uuid.New()
	platformID := uuid.New()
	integrationID := uuid.New()
	integrationRepo.integrations[integrationID.String()] = sqlc.Integration{
		ID:                integrationID,
		ClientID:          clientID,
		PlatformID:        platformID,
		ExternalAccountID: "loc-123",
		Status:            sqlc.IntegrationStatusACTIVE,
	}
	// Pre-existing config for the integration.
	configRepo.configs["existing"] = sqlc.PaymentProviderConfig{
		ID:            uuid.New(),
		IntegrationID: integrationID,
		LocationID:    "loc-123",
	}

	dispatcher := NewHighLevelWebhookDispatcher(
		&testDispatcherLogger{},
		integrationRepo,
		configRepo,
		ProviderConfigSettings{},
	)

	event := &WebhookEvent{
		Provider:        "highlevel",
		EventType:       "INSTALL",
		ProviderEventID: "evt-1",
		LocationID:      "loc-123",
	}

	err := dispatcher.Dispatch(context.Background(), event)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	// The existing config must be reused (not duplicated).
	if len(configRepo.configs) != 1 {
		t.Fatalf("expected 1 config (reused), got %d", len(configRepo.configs))
	}
}

func TestDispatch_InstallMultipleClientsSelectsCorrect(t *testing.T) {
	t.Parallel()

	integrationRepo := newTestDispatcherIntegrationRepo()
	configRepo := newTestDispatcherConfigRepo()

	// Client A -> Integration A (location loc-A)
	clientA := uuid.New()
	platformID := uuid.New()
	integrationA := uuid.New()
	integrationRepo.integrations[integrationA.String()] = sqlc.Integration{
		ID:                integrationA,
		ClientID:          clientA,
		PlatformID:        platformID,
		ExternalAccountID: "loc-A",
		Status:            sqlc.IntegrationStatusACTIVE,
	}

	// Client B -> Integration B (location loc-B)
	clientB := uuid.New()
	integrationB := uuid.New()
	integrationRepo.integrations[integrationB.String()] = sqlc.Integration{
		ID:                integrationB,
		ClientID:          clientB,
		PlatformID:        platformID,
		ExternalAccountID: "loc-B",
		Status:            sqlc.IntegrationStatusACTIVE,
	}

	dispatcher := NewHighLevelWebhookDispatcher(
		&testDispatcherLogger{},
		integrationRepo,
		configRepo,
		ProviderConfigSettings{},
	)

	// INSTALL for Client B's location must resolve to Integration B, not A.
	event := &WebhookEvent{
		Provider:        "highlevel",
		EventType:       "INSTALL",
		ProviderEventID: "evt-B",
		LocationID:      "loc-B",
	}

	err := dispatcher.Dispatch(context.Background(), event)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if len(configRepo.configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configRepo.configs))
	}
	for _, c := range configRepo.configs {
		if c.IntegrationID != integrationB {
			t.Fatalf("config IntegrationID = %v, want %v (Client B's integration)", c.IntegrationID, integrationB)
		}
		if c.LocationID != "loc-B" {
			t.Fatalf("config LocationID = %q, want loc-B", c.LocationID)
		}
	}
}

func TestDispatch_InstallMissingMappingFailsSafely(t *testing.T) {
	t.Parallel()

	integrationRepo := newTestDispatcherIntegrationRepo()
	configRepo := newTestDispatcherConfigRepo()

	dispatcher := NewHighLevelWebhookDispatcher(
		&testDispatcherLogger{},
		integrationRepo,
		configRepo,
		ProviderConfigSettings{},
	)

	// No integration or config exists for this location.
	event := &WebhookEvent{
		Provider:        "highlevel",
		EventType:       "INSTALL",
		ProviderEventID: "evt-1",
		LocationID:      "unknown-loc",
	}

	err := dispatcher.Dispatch(context.Background(), event)
	if err == nil {
		t.Fatal("Dispatch should fail when no integration can be resolved")
	}

	// No config should be created.
	if len(configRepo.configs) != 0 {
		t.Fatalf("expected 0 configs, got %d", len(configRepo.configs))
	}
}

func TestDispatch_InstallMissingLocationID(t *testing.T) {
	t.Parallel()

	integrationRepo := newTestDispatcherIntegrationRepo()
	configRepo := newTestDispatcherConfigRepo()

	dispatcher := NewHighLevelWebhookDispatcher(
		&testDispatcherLogger{},
		integrationRepo,
		configRepo,
		ProviderConfigSettings{},
	)

	event := &WebhookEvent{
		Provider:        "highlevel",
		EventType:       "INSTALL",
		ProviderEventID: "evt-1",
		LocationID:      "",
	}

	err := dispatcher.Dispatch(context.Background(), event)
	if err == nil {
		t.Fatal("Dispatch should fail when locationId is missing")
	}
}
