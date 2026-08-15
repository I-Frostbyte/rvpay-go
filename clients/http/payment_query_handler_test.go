package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/I-Frostbyte/rvpay-go/clients/payments"
	transactionsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockConfigRepo is a minimal PaymentProviderConfigRepo test double.
type mockConfigRepo struct {
	configs map[string]sqlc.PaymentProviderConfig
}

func newMockConfigRepo() *mockConfigRepo {
	return &mockConfigRepo{configs: make(map[string]sqlc.PaymentProviderConfig)}
}

func (m *mockConfigRepo) Create(ctx context.Context, integrationID uuid.UUID, providerName, providerDescription, providerImageURL, locationID, queryURL, paymentsURL string, supportsSubscriptionSchedule bool, providerAPIKey string) (sqlc.PaymentProviderConfig, error) {
	return sqlc.PaymentProviderConfig{}, nil
}

func (m *mockConfigRepo) GetByIntegrationID(ctx context.Context, integrationID uuid.UUID) (sqlc.PaymentProviderConfig, error) {
	c, ok := m.configs[integrationID.String()]
	if !ok {
		return sqlc.PaymentProviderConfig{}, repo.ErrNotFound
	}
	return c, nil
}

func (m *mockConfigRepo) GetByLocationID(ctx context.Context, locationID string) (sqlc.PaymentProviderConfig, error) {
	for _, c := range m.configs {
		if c.LocationID == locationID {
			return c, nil
		}
	}
	return sqlc.PaymentProviderConfig{}, repo.ErrNotFound
}

func (m *mockConfigRepo) GetByAPIKey(ctx context.Context, apiKey string) (sqlc.PaymentProviderConfig, error) {
	for _, c := range m.configs {
		if c.ProviderApiKey == apiKey {
			return c, nil
		}
	}
	return sqlc.PaymentProviderConfig{}, repo.ErrNotFound
}

func (m *mockConfigRepo) Update(ctx context.Context, integrationID uuid.UUID, providerName, providerDescription, providerImageURL, locationID, queryURL, paymentsURL string, supportsSubscriptionSchedule bool, providerAPIKey string) (sqlc.PaymentProviderConfig, error) {
	return sqlc.PaymentProviderConfig{}, nil
}

func (m *mockConfigRepo) Delete(ctx context.Context, integrationID uuid.UUID) error {
	return nil
}

// mockIntegrationRepo is a minimal IntegrationRepo test double.
type mockIntegrationRepo struct {
	integrations map[string]sqlc.Integration
}

func newMockIntegrationRepo() *mockIntegrationRepo {
	return &mockIntegrationRepo{integrations: make(map[string]sqlc.Integration)}
}

func (m *mockIntegrationRepo) Create(ctx context.Context, clientID, platformID uuid.UUID, externalAccountID string, status sqlc.IntegrationStatus) (sqlc.Integration, error) {
	return sqlc.Integration{}, nil
}

func (m *mockIntegrationRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Integration, error) {
	i, ok := m.integrations[id.String()]
	if !ok {
		return sqlc.Integration{}, repo.ErrNotFound
	}
	return i, nil
}

func (m *mockIntegrationRepo) GetByClientAndPlatform(ctx context.Context, clientID, platformID uuid.UUID) (sqlc.Integration, error) {
	return sqlc.Integration{}, repo.ErrNotFound
}

func (m *mockIntegrationRepo) GetByExternalAccountID(ctx context.Context, externalAccountID string) (sqlc.Integration, error) {
	return sqlc.Integration{}, repo.ErrNotFound
}

func (m *mockIntegrationRepo) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	return nil, nil
}

func (m *mockIntegrationRepo) ListByPlatform(ctx context.Context, platformID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	return nil, nil
}

func (m *mockIntegrationRepo) ListActiveByClient(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	return nil, nil
}

func (m *mockIntegrationRepo) CountByClient(ctx context.Context, clientID uuid.UUID) (int64, error) {
	return 0, nil
}

func (m *mockIntegrationRepo) ExistsByClientAndPlatform(ctx context.Context, clientID, platformID uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockIntegrationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.IntegrationStatus) (sqlc.Integration, error) {
	return sqlc.Integration{}, nil
}

func (m *mockIntegrationRepo) UpdateLastSyncAt(ctx context.Context, id uuid.UUID, lastSyncAt pgtype.Timestamptz) (sqlc.Integration, error) {
	return sqlc.Integration{}, nil
}

func (m *mockIntegrationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// mockWebhookEventRepo is a minimal WebhookEventRepo test double.
type mockWebhookEventRepo struct {
	events map[string]bool
}

func newMockWebhookEventRepo() *mockWebhookEventRepo {
	return &mockWebhookEventRepo{events: make(map[string]bool)}
}

func (m *mockWebhookEventRepo) Create(ctx context.Context, integrationID uuid.UUID, providerEventID, eventType string, payload []byte) (sqlc.WebhookEvent, error) {
	key := integrationID.String() + ":" + providerEventID
	if m.events[key] {
		return sqlc.WebhookEvent{}, repo.ErrDuplicate
	}
	m.events[key] = true
	return sqlc.WebhookEvent{}, nil
}

func (m *mockWebhookEventRepo) GetByIntegrationAndProvider(ctx context.Context, integrationID uuid.UUID, providerEventID string) (sqlc.WebhookEvent, error) {
	return sqlc.WebhookEvent{}, repo.ErrNotFound
}

// fakeTransactionsClient is a fake PaymentServiceClient test double.
type fakeTransactionsClient struct {
	verifyResults map[string]*transactionsgrpc.VerifyPaymentResponse
}

func newFakeTransactionsClient() *fakeTransactionsClient {
	return &fakeTransactionsClient{verifyResults: make(map[string]*transactionsgrpc.VerifyPaymentResponse)}
}

func (f *fakeTransactionsClient) VerifyPayment(ctx context.Context, in *transactionsgrpc.VerifyPaymentRequest, opts ...grpc.CallOption) (*transactionsgrpc.VerifyPaymentResponse, error) {
	// RVPay only supports one-time payments; subscription-scoped verification
	// is rejected by the Transactions service.
	if in.GetSubscriptionId() != "" {
		return nil, status.Error(codes.FailedPrecondition, "subscription payments are not supported")
	}
	// Try transaction ID first, then charge ID.
	if resp, ok := f.verifyResults[in.GetGhlTransactionId()]; ok {
		return resp, nil
	}
	if resp, ok := f.verifyResults[in.GetGhlChargeId()]; ok {
		return resp, nil
	}
	return nil, status.Error(codes.NotFound, "transaction not found")
}

func (f *fakeTransactionsClient) ProcessPaymentWebhook(ctx context.Context, in *transactionsgrpc.ProcessPaymentWebhookRequest, opts ...grpc.CallOption) (*transactionsgrpc.ProcessPaymentWebhookResponse, error) {
	return &transactionsgrpc.ProcessPaymentWebhookResponse{}, nil
}

func newTestPaymentQueryHandler() (*PaymentQueryHandler, *mockConfigRepo, *fakeTransactionsClient) {
	configRepo := newMockConfigRepo()
	integrationRepo := newMockIntegrationRepo()
	webhookEventRepo := newMockWebhookEventRepo()
	transactionsClient := newFakeTransactionsClient()
	logger := zerolog.Nop()

	svc := payments.NewService(configRepo, integrationRepo, webhookEventRepo, transactionsClient, logger)
	handler := NewPaymentQueryHandler(svc, logger)
	return handler, configRepo, transactionsClient
}

func TestPaymentQueryHandler_UnsupportedMethod(t *testing.T) {
	handler, _, _ := newTestPaymentQueryHandler()

	req := httptest.NewRequest(http.MethodGet, "/payments/custom-provider/query", nil)
	rec := httptest.NewRecorder()

	handler.Query(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestPaymentQueryHandler_MalformedRequest(t *testing.T) {
	handler, _, _ := newTestPaymentQueryHandler()

	req := httptest.NewRequest(http.MethodPost, "/payments/custom-provider/query", bytes.NewBufferString("not-json"))
	rec := httptest.NewRecorder()

	handler.Query(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPaymentQueryHandler_MissingAPIKey(t *testing.T) {
	handler, _, _ := newTestPaymentQueryHandler()

	body, _ := json.Marshal(payments.QueryRequest{
		Type:          "verify",
		TransactionID: "txn-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/payments/custom-provider/query", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.Query(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPaymentQueryHandler_InvalidAPIKey(t *testing.T) {
	handler, _, _ := newTestPaymentQueryHandler()

	body, _ := json.Marshal(payments.QueryRequest{
		Type:          "verify",
		TransactionID: "txn-1",
		APIKey:        "wrong-key",
	})
	req := httptest.NewRequest(http.MethodPost, "/payments/custom-provider/query", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.Query(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestPaymentQueryHandler_UnsupportedType(t *testing.T) {
	handler, configRepo, _ := newTestPaymentQueryHandler()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}

	body, _ := json.Marshal(payments.QueryRequest{
		Type:          "refund",
		TransactionID: "txn-1",
		APIKey:        "valid-key",
	})
	req := httptest.NewRequest(http.MethodPost, "/payments/custom-provider/query", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.Query(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPaymentQueryHandler_VerifySuccess(t *testing.T) {
	handler, configRepo, txClient := newTestPaymentQueryHandler()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}
	txClient.verifyResults["txn-1"] = &transactionsgrpc.VerifyPaymentResponse{Success: true}

	body, _ := json.Marshal(payments.QueryRequest{
		Type:          "verify",
		TransactionID: "txn-1",
		APIKey:        "valid-key",
	})
	req := httptest.NewRequest(http.MethodPost, "/payments/custom-provider/query", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.Query(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp payments.QueryResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected success=true in response")
	}
}

func TestPaymentQueryHandler_VerifyFailed(t *testing.T) {
	handler, configRepo, txClient := newTestPaymentQueryHandler()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}
	txClient.verifyResults["txn-1"] = &transactionsgrpc.VerifyPaymentResponse{Failed: true}

	body, _ := json.Marshal(payments.QueryRequest{
		Type:          "verify",
		TransactionID: "txn-1",
		APIKey:        "valid-key",
	})
	req := httptest.NewRequest(http.MethodPost, "/payments/custom-provider/query", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.Query(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp payments.QueryResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Failed {
		t.Fatal("expected failed=true in response")
	}
}

func TestPaymentQueryHandler_VerifyPending(t *testing.T) {
	handler, configRepo, txClient := newTestPaymentQueryHandler()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}
	txClient.verifyResults["txn-1"] = &transactionsgrpc.VerifyPaymentResponse{}

	body, _ := json.Marshal(payments.QueryRequest{
		Type:          "verify",
		TransactionID: "txn-1",
		APIKey:        "valid-key",
	})
	req := httptest.NewRequest(http.MethodPost, "/payments/custom-provider/query", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.Query(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp payments.QueryResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Success {
		t.Fatal("expected success=false for pending")
	}
	if resp.Failed {
		t.Fatal("expected failed=false for pending")
	}
}

func TestPaymentQueryHandler_VerifyUnknownTransaction(t *testing.T) {
	handler, configRepo, _ := newTestPaymentQueryHandler()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}

	body, _ := json.Marshal(payments.QueryRequest{
		Type:          "verify",
		TransactionID: "unknown-txn",
		APIKey:        "valid-key",
	})
	req := httptest.NewRequest(http.MethodPost, "/payments/custom-provider/query", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.Query(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestPaymentQueryHandler_VerifyByChargeID(t *testing.T) {
	handler, configRepo, txClient := newTestPaymentQueryHandler()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}
	txClient.verifyResults["charge-1"] = &transactionsgrpc.VerifyPaymentResponse{Success: true}

	// HighLevel may send only a chargeId (no transactionId).
	body, _ := json.Marshal(payments.QueryRequest{
		Type:     "verify",
		APIKey:   "valid-key",
		ChargeID: "charge-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/payments/custom-provider/query", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.Query(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp payments.QueryResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected success=true for charge ID lookup")
	}
}

func TestPaymentQueryHandler_VerifySubscriptionRejected(t *testing.T) {
	handler, configRepo, _ := newTestPaymentQueryHandler()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}

	body, _ := json.Marshal(payments.QueryRequest{
		Type:           "verify",
		TransactionID:  "txn-1",
		APIKey:         "valid-key",
		SubscriptionID: "sub-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/payments/custom-provider/query", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.Query(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
