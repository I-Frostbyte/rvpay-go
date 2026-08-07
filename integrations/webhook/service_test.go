package webhook

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/I-Frostbyte/rvpay-go/integrations/db/repo/mocks"
	"github.com/I-Frostbyte/rvpay-go/integrations/db/sqlc"
	sqlcmocks "github.com/I-Frostbyte/rvpay-go/integrations/db/sqlc/mocks"
	"github.com/rs/zerolog"
	"go.uber.org/mock/gomock"
)

func newTestService(t *testing.T, repo *mocks.MockIntegrationsRepo) *Service {
	t.Helper()
	return NewService(repo, zerolog.Nop())
}

func TestHandleWebhookRejectsEmptyBody(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	svc := newTestService(t, repo)

	err := svc.HandleWebhook(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for empty body, got nil")
	}
}

func TestHandleWebhookRejectsMalformedPayload(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	svc := newTestService(t, repo)

	err := svc.HandleWebhook(context.Background(), []byte(`not-json`))
	if err == nil {
		t.Fatal("expected error for malformed payload, got nil")
	}
	if !strings.Contains(err.Error(), "invalid webhook payload") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestHandleWebhookRejectsMissingEventType(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	svc := newTestService(t, repo)

	err := svc.HandleWebhook(context.Background(), []byte(`{"data":{}}`))
	if err == nil {
		t.Fatal("expected error for missing event type, got nil")
	}
	if !strings.Contains(err.Error(), "webhook event type is required") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestHandleWebhookStoreFailure(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	querier := sqlcmocks.NewMockQuerier(ctrl)
	querier.EXPECT().CreateWebhookEvent(gomock.Any(), gomock.Any()).
		Return(sqlc.WebhookEvent{}, errors.New("db error"))

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	repo.EXPECT().Do().Return(querier)

	svc := newTestService(t, repo)

	err := svc.HandleWebhook(context.Background(), []byte(`{"type":"ContactCreate","eventType":"contact.created","data":{}}`))
	if err == nil {
		t.Fatal("expected error for store failure, got nil")
	}
	if !strings.Contains(err.Error(), "could not store webhook event") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestHandleWebhookSuccessWithEventType(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	querier := sqlcmocks.NewMockQuerier(ctrl)
	querier.EXPECT().CreateWebhookEvent(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params sqlc.CreateWebhookEventParams) (sqlc.WebhookEvent, error) {
			if params.Provider != "highlevel" {
				t.Fatalf("provider = %s, want highlevel", params.Provider)
			}
			if params.EventType != "contact.created" {
				t.Fatalf("event_type = %s, want contact.created", params.EventType)
			}
			if params.Processed {
				t.Fatal("processed should default to false")
			}
			if len(params.Payload) == 0 {
				t.Fatal("payload was not stored")
			}
			return sqlc.WebhookEvent{}, nil
		})

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	repo.EXPECT().Do().Return(querier)

	svc := newTestService(t, repo)

	err := svc.HandleWebhook(context.Background(), []byte(`{"type":"ContactCreate","eventType":"contact.created","data":{"id":"contact_1"}}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleWebhookSuccessFallsBackToType(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	querier := sqlcmocks.NewMockQuerier(ctrl)
	querier.EXPECT().CreateWebhookEvent(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params sqlc.CreateWebhookEventParams) (sqlc.WebhookEvent, error) {
			if params.EventType != "ContactCreate" {
				t.Fatalf("event_type = %s, want ContactCreate (fallback to type)", params.EventType)
			}
			return sqlc.WebhookEvent{}, nil
		})

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	repo.EXPECT().Do().Return(querier)

	svc := newTestService(t, repo)

	err := svc.HandleWebhook(context.Background(), []byte(`{"type":"ContactCreate","data":{"id":"contact_1"}}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadBody(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	svc := newTestService(t, repo)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/highlevel", strings.NewReader(`{"eventType":"contact.created"}`))
	body, err := svc.ReadBody(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(body) != `{"eventType":"contact.created"}` {
		t.Fatalf("body = %s, want %s", string(body), `{"eventType":"contact.created"}`)
	}
}

func TestHandlerHighLevelMethodNotAllowed(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	svc := newTestService(t, repo)
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/webhooks/highlevel", nil)
	rec := httptest.NewRecorder()

	handler.HighLevel(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandlerHighLevelBadRequest(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	svc := newTestService(t, repo)
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/highlevel", strings.NewReader(`not-json`))
	rec := httptest.NewRecorder()

	handler.HighLevel(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandlerHighLevelSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	querier := sqlcmocks.NewMockQuerier(ctrl)
	querier.EXPECT().CreateWebhookEvent(gomock.Any(), gomock.Any()).
		Return(sqlc.WebhookEvent{}, nil)

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	repo.EXPECT().Do().Return(querier)

	svc := newTestService(t, repo)
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/highlevel", strings.NewReader(`{"eventType":"contact.created","data":{}}`))
	rec := httptest.NewRecorder()

	handler.HighLevel(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("body = %s, want ok", rec.Body.String())
	}
}