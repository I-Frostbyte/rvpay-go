package payments

import (
	"context"
	"sync"
	"testing"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
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

// mockWebhookEventRepo is a minimal WebhookEventRepo test double with a mutex
// for safe concurrent access in race-detected tests.
type mockWebhookEventRepo struct {
	mu     sync.Mutex
	events map[string]bool
}

func newMockWebhookEventRepo() *mockWebhookEventRepo {
	return &mockWebhookEventRepo{events: make(map[string]bool)}
}

func (m *mockWebhookEventRepo) Create(ctx context.Context, integrationID uuid.UUID, providerEventID, eventType string, payload []byte) (sqlc.WebhookEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
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
	// verifyResults maps a GHL transaction ID to the verification response.
	verifyResults map[string]*transactionsgrpc.VerifyPaymentResponse
	// webhookProcessed records whether ProcessPaymentWebhook was called.
	webhookProcessed bool
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
	f.webhookProcessed = true
	return &transactionsgrpc.ProcessPaymentWebhookResponse{}, nil
}

func newTestService() (*Service, *mockConfigRepo, *mockIntegrationRepo, *mockWebhookEventRepo, *fakeTransactionsClient) {
	configRepo := newMockConfigRepo()
	integrationRepo := newMockIntegrationRepo()
	webhookEventRepo := newMockWebhookEventRepo()
	transactionsClient := newFakeTransactionsClient()
	logger := zerolog.Nop()

	svc := NewService(configRepo, integrationRepo, webhookEventRepo, transactionsClient, logger)
	return svc, configRepo, integrationRepo, webhookEventRepo, transactionsClient
}

func TestHandleQuery_MissingAPIKey(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	_, err := svc.HandleQuery(context.Background(), &QueryRequest{Type: "verify", TransactionID: "txn-1"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", status.Code(err))
	}
}

func TestHandleQuery_InvalidAPIKey(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	_, err := svc.HandleQuery(context.Background(), &QueryRequest{Type: "verify", TransactionID: "txn-1", APIKey: "wrong-key"})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestHandleQuery_UnsupportedType(t *testing.T) {
	svc, configRepo, _, _, _ := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}

	_, err := svc.HandleQuery(context.Background(), &QueryRequest{Type: "refund", TransactionID: "txn-1", APIKey: "valid-key"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", status.Code(err))
	}
}

func TestHandleQuery_MissingTransactionID(t *testing.T) {
	svc, configRepo, _, _, _ := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}

	_, err := svc.HandleQuery(context.Background(), &QueryRequest{Type: "verify", APIKey: "valid-key"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", status.Code(err))
	}
}

func TestHandleQuery_Verify_Completed(t *testing.T) {
	svc, configRepo, _, _, txClient := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}
	txClient.verifyResults["txn-1"] = &transactionsgrpc.VerifyPaymentResponse{Success: true}

	resp, err := svc.HandleQuery(context.Background(), &QueryRequest{Type: "verify", TransactionID: "txn-1", APIKey: "valid-key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success=true for completed deposit")
	}
	if resp.Failed {
		t.Fatalf("expected failed=false for completed deposit")
	}
}

func TestHandleQuery_Verify_Failed(t *testing.T) {
	svc, configRepo, _, _, txClient := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}
	txClient.verifyResults["txn-1"] = &transactionsgrpc.VerifyPaymentResponse{Failed: true}

	resp, err := svc.HandleQuery(context.Background(), &QueryRequest{Type: "verify", TransactionID: "txn-1", APIKey: "valid-key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Failed {
		t.Fatalf("expected failed=true for failed deposit")
	}
	if resp.Success {
		t.Fatalf("expected success=false for failed deposit")
	}
}

func TestHandleQuery_Verify_Pending(t *testing.T) {
	svc, configRepo, _, _, txClient := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}
	txClient.verifyResults["txn-1"] = &transactionsgrpc.VerifyPaymentResponse{}

	resp, err := svc.HandleQuery(context.Background(), &QueryRequest{Type: "verify", TransactionID: "txn-1", APIKey: "valid-key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatalf("expected success=false for pending deposit")
	}
	if resp.Failed {
		t.Fatalf("expected failed=false for pending deposit")
	}
}

func TestHandleQuery_Verify_ByChargeID(t *testing.T) {
	svc, configRepo, _, _, txClient := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}
	txClient.verifyResults["charge-1"] = &transactionsgrpc.VerifyPaymentResponse{Success: true}

	// HighLevel may send only a chargeId (no transactionId).
	resp, err := svc.HandleQuery(context.Background(), &QueryRequest{Type: "verify", APIKey: "valid-key", ChargeID: "charge-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success=true for charge ID lookup")
	}
}

func TestHandleQuery_Verify_SubscriptionRejected(t *testing.T) {
	svc, configRepo, _, _, _ := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}

	// Subscription-scoped verification is rejected by the Transactions service.
	_, err := svc.HandleQuery(context.Background(), &QueryRequest{Type: "verify", TransactionID: "txn-1", APIKey: "valid-key", SubscriptionID: "sub-1"})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v", status.Code(err))
	}
}

func TestHandleQuery_Verify_UnknownTransaction(t *testing.T) {
	svc, configRepo, _, _, _ := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}

	_, err := svc.HandleQuery(context.Background(), &QueryRequest{Type: "verify", TransactionID: "unknown-txn", APIKey: "valid-key"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", status.Code(err))
	}
}

func TestHandleWebhook_InvalidPayload(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	err := svc.HandleWebhook(context.Background(), []byte("not-json"))
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", status.Code(err))
	}
}

func TestHandleWebhook_MissingEventID(t *testing.T) {
	svc, configRepo, _, _, _ := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}

	err := svc.HandleWebhook(context.Background(), []byte(`{"eventType":"payment.captured","locationId":"loc-1","apiKey":"valid-key"}`))
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", status.Code(err))
	}
}

func TestHandleWebhook_UnknownLocation(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	err := svc.HandleWebhook(context.Background(), []byte(`{"eventType":"payment.captured","eventId":"evt-1","locationId":"unknown-loc","apiKey":"some-key"}`))
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", status.Code(err))
	}
}

func TestHandleWebhook_PaymentCaptured(t *testing.T) {
	svc, configRepo, integrationRepo, _, txClient := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}
	integrationRepo.integrations[integrationID.String()] = sqlc.Integration{
		ID:     integrationID,
		Status: sqlc.IntegrationStatusACTIVE,
	}
	txClient.verifyResults["txn-1"] = &transactionsgrpc.VerifyPaymentResponse{Success: true}

	err := svc.HandleWebhook(context.Background(), []byte(`{"eventType":"payment.captured","eventId":"evt-1","locationId":"loc-1","transactionId":"txn-1","chargeId":"charge-1","apiKey":"valid-key"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleWebhook_DuplicateEvent(t *testing.T) {
	svc, configRepo, integrationRepo, _, txClient := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}
	integrationRepo.integrations[integrationID.String()] = sqlc.Integration{
		ID:     integrationID,
		Status: sqlc.IntegrationStatusACTIVE,
	}
	txClient.verifyResults["txn-1"] = &transactionsgrpc.VerifyPaymentResponse{Success: true}

	body := []byte(`{"eventType":"payment.captured","eventId":"evt-1","locationId":"loc-1","transactionId":"txn-1","chargeId":"charge-1","apiKey":"valid-key"}`)

	// First delivery succeeds.
	if err := svc.HandleWebhook(context.Background(), body); err != nil {
		t.Fatalf("first delivery failed: %v", err)
	}

	// Duplicate delivery is acknowledged as a duplicate.
	err := svc.HandleWebhook(context.Background(), body)
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("expected AlreadyExists for duplicate, got %v", status.Code(err))
	}
}

func TestHandleWebhook_MissingAPIKey(t *testing.T) {
	svc, configRepo, _, _, _ := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}

	err := svc.HandleWebhook(context.Background(), []byte(`{"eventType":"payment.captured","eventId":"evt-1","locationId":"loc-1"}`))
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for missing API key, got %v", status.Code(err))
	}
}

func TestHandleWebhook_InvalidAPIKey(t *testing.T) {
	svc, configRepo, _, _, _ := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}

	err := svc.HandleWebhook(context.Background(), []byte(`{"eventType":"payment.captured","eventId":"evt-1","locationId":"loc-1","apiKey":"wrong-key"}`))
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated for invalid API key, got %v", status.Code(err))
	}
}

func TestHandleWebhook_ThreeIdenticalEvents(t *testing.T) {
	svc, configRepo, integrationRepo, _, txClient := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}
	integrationRepo.integrations[integrationID.String()] = sqlc.Integration{
		ID:     integrationID,
		Status: sqlc.IntegrationStatusACTIVE,
	}
	txClient.verifyResults["txn-1"] = &transactionsgrpc.VerifyPaymentResponse{Success: true}

	body := []byte(`{"eventType":"payment.captured","eventId":"evt-1","locationId":"loc-1","transactionId":"txn-1","chargeId":"charge-1","apiKey":"valid-key"}`)

	// First delivery succeeds.
	if err := svc.HandleWebhook(context.Background(), body); err != nil {
		t.Fatalf("first delivery failed: %v", err)
	}

	// Second delivery is duplicate.
	err := svc.HandleWebhook(context.Background(), body)
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("expected AlreadyExists for second delivery, got %v", status.Code(err))
	}

	// Third delivery is also duplicate.
	err = svc.HandleWebhook(context.Background(), body)
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("expected AlreadyExists for third delivery, got %v", status.Code(err))
	}
}

func TestHandleWebhook_UnknownEventType(t *testing.T) {
	svc, configRepo, integrationRepo, _, _ := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}
	integrationRepo.integrations[integrationID.String()] = sqlc.Integration{
		ID:     integrationID,
		Status: sqlc.IntegrationStatusACTIVE,
	}

	// Unknown event types are acknowledged safely without processing.
	err := svc.HandleWebhook(context.Background(), []byte(`{"eventType":"subscription.active","eventId":"evt-2","locationId":"loc-1","apiKey":"valid-key"}`))
	if err != nil {
		t.Fatalf("unexpected error for unknown event type: %v", err)
	}
}

func TestHandleWebhook_WrongChargeID(t *testing.T) {
	svc, configRepo, integrationRepo, _, txClient := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}
	integrationRepo.integrations[integrationID.String()] = sqlc.Integration{
		ID:     integrationID,
		Status: sqlc.IntegrationStatusACTIVE,
	}
	txClient.verifyResults["txn-1"] = &transactionsgrpc.VerifyPaymentResponse{Success: true}

	// A payment.captured event with a transaction ID that resolves but a
	// charge ID that doesn't match the deposit is still processed; the
	// charge ID is recorded as a reference. This is not an error condition.
	err := svc.HandleWebhook(context.Background(), []byte(`{"eventType":"payment.captured","eventId":"evt-1","locationId":"loc-1","transactionId":"txn-1","chargeId":"wrong-charge","apiKey":"valid-key"}`))
	if err != nil {
		t.Fatalf("unexpected error for wrong charge ID: %v", err)
	}
}

// TestHandleWebhook_ConcurrentDuplicateEvents verifies that duplicate
// webhook deliveries arriving concurrently are handled safely. Only one
// should succeed; the rest must be acknowledged as duplicates without
// corrupting state.
func TestHandleWebhook_ConcurrentDuplicateEvents(t *testing.T) {
	svc, configRepo, integrationRepo, _, txClient := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}
	integrationRepo.integrations[integrationID.String()] = sqlc.Integration{
		ID:     integrationID,
		Status: sqlc.IntegrationStatusACTIVE,
	}
	txClient.verifyResults["txn-1"] = &transactionsgrpc.VerifyPaymentResponse{Success: true}

	body := []byte(`{"eventType":"payment.captured","eventId":"evt-concurrent","locationId":"loc-1","transactionId":"txn-1","chargeId":"charge-1","apiKey":"valid-key"}`)

	// Launch 10 concurrent goroutines, all delivering the same event.
	var wg sync.WaitGroup
	errs := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- svc.HandleWebhook(context.Background(), body)
		}()
	}
	wg.Wait()
	close(errs)

	// Exactly one should succeed; the rest should be AlreadyExists.
	successCount := 0
	dupCount := 0
	for err := range errs {
		if err == nil {
			successCount++
		} else if status.Code(err) == codes.AlreadyExists {
			dupCount++
		} else {
			t.Errorf("unexpected error code: %v", status.Code(err))
		}
	}
	if successCount != 1 {
		t.Errorf("expected exactly 1 success, got %d", successCount)
	}
	if dupCount != 9 {
		t.Errorf("expected exactly 9 duplicates, got %d", dupCount)
	}
}

// TestHandleWebhook_ConcurrentSameIDempotentKey verifies that when both
// eventID and integrationID are the same, only one call succeeds.
func TestHandleWebhook_ConcurrentSameIDempotentKey(t *testing.T) {
	svc, configRepo, integrationRepo, _, txClient := newTestService()
	integrationID := uuid.New()
	configRepo.configs[integrationID.String()] = sqlc.PaymentProviderConfig{
		IntegrationID:  integrationID,
		LocationID:     "loc-1",
		ProviderApiKey: "valid-key",
	}
	integrationRepo.integrations[integrationID.String()] = sqlc.Integration{
		ID:     integrationID,
		Status: sqlc.IntegrationStatusACTIVE,
	}
	txClient.verifyResults["txn-1"] = &transactionsgrpc.VerifyPaymentResponse{Success: true}

	body := []byte(`{"eventType":"payment.captured","eventId":"evt-same-id","locationId":"loc-1","transactionId":"txn-1","chargeId":"charge-1","apiKey":"valid-key"}`)

	var wg sync.WaitGroup
	errCh := make(chan error, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errCh <- svc.HandleWebhook(context.Background(), body)
		}()
	}
	wg.Wait()
	close(errCh)

	successCount := 0
	dupCount := 0
	for err := range errCh {
		if err == nil {
			successCount++
		} else if status.Code(err) == codes.AlreadyExists {
			dupCount++
		} else {
			t.Errorf("unexpected error: %v", err)
		}
	}
	if successCount != 1 {
		t.Errorf("expected exactly 1 success, got %d", successCount)
	}
	if dupCount != 4 {
		t.Errorf("expected exactly 4 duplicates, got %d", dupCount)
	}
}

// TestHandleQuery_NilRequest verifies that passing a nil query request
// returns an InvalidArgument error.
func TestHandleQuery_NilRequest(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	_, err := svc.HandleQuery(context.Background(), nil)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for nil request, got %v", status.Code(err))
	}
}
