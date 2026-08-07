package oauth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/I-Frostbyte/rvpay-go/integrations/db/repo/mocks"
	"github.com/I-Frostbyte/rvpay-go/integrations/db/sqlc"
	sqlcmocks "github.com/I-Frostbyte/rvpay-go/integrations/db/sqlc/mocks"
	"go.uber.org/mock/gomock"
)

func TestHandlerCallbackMethodNotAllowed(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	svc := newTestService(t, repo, &mockRoundTripper{})
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/oauth/callback", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandlerCallbackMissingCode(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	svc := newTestService(t, repo, &mockRoundTripper{})
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/oauth/callback", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), "authorization code is required") {
		t.Fatalf("body = %s, want authorization code error", rec.Body.String())
	}
}

func TestHandlerCallbackServiceError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	svc := newTestService(t, repo, &mockRoundTripper{
		statusCode: http.StatusBadRequest,
		body:       `{"error":"invalid_grant"}`,
	})
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?code=bad-code", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestHandlerCallbackSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	querier := sqlcmocks.NewMockQuerier(ctrl)
	querier.EXPECT().CreateIntegration(gomock.Any(), gomock.Any()).
		Return(sqlc.Integration{}, nil)

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	repo.EXPECT().Do().Return(querier)

	svc := newTestService(t, repo, &mockRoundTripper{
		statusCode: http.StatusOK,
		body: `{
			"access_token": "access-token",
			"refresh_token": "refresh-token",
			"expires_in": 3600,
			"token_type": "Bearer",
			"scope": "read",
			"userType": "Company",
			"companyId": "company_1",
			"locationId": "loc_123",
			"userId": "user_1",
			"planId": "plan_1"
		}`,
	})
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?code=valid-code", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "integration authorized successfully" {
		t.Fatalf("body = %s, want integration authorized successfully", rec.Body.String())
	}
}

func TestHandlerCallbackServicePanic(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	svc := newTestService(t, repo, &mockRoundTripper{
		err: errors.New("network down"),
	})
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?code=valid-code", nil)
	rec := httptest.NewRecorder()

	handler.Callback(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}