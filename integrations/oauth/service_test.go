package oauth

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/I-Frostbyte/rvpay-go/integrations/db/repo/mocks"
	"github.com/I-Frostbyte/rvpay-go/integrations/db/sqlc"
	sqlcmocks "github.com/I-Frostbyte/rvpay-go/integrations/db/sqlc/mocks"
	"github.com/rs/zerolog"
	"go.uber.org/mock/gomock"
)

// mockRoundTripper intercepts HTTP requests and returns a canned response.
type mockRoundTripper struct {
	statusCode int
	body       string
	err        error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(strings.NewReader(m.body)),
		Header:     make(http.Header),
	}, nil
}

func newTestService(t *testing.T, repo *mocks.MockIntegrationsRepo, rt http.RoundTripper) *Service {
	t.Helper()

	svc := NewService(
		repo,
		zerolog.Nop(),
		"test-client-id",
		"test-client-secret",
		"http://localhost:8080/oauth/callback",
		"0123456789abcdef0123456789abcdef", // 32-byte AES key
	)
	svc.httpClient = &http.Client{Transport: rt}
	return svc
}

func TestHandleCallbackRejectsEmptyCode(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	svc := newTestService(t, repo, &mockRoundTripper{})

	err := svc.HandleCallback(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty code, got nil")
	}
}

func TestHandleCallbackExchangeFailure(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	svc := newTestService(t, repo, &mockRoundTripper{
		statusCode: http.StatusBadRequest,
		body:       `{"error":"invalid_grant"}`,
	})

	err := svc.HandleCallback(context.Background(), "some-code")
	if err == nil {
		t.Fatal("expected error for failed token exchange, got nil")
	}
	if !strings.Contains(err.Error(), "highlevel returned status 400") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestHandleCallbackTransportError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	svc := newTestService(t, repo, &mockRoundTripper{
		err: errors.New("network down"),
	})

	err := svc.HandleCallback(context.Background(), "some-code")
	if err == nil {
		t.Fatal("expected error for transport failure, got nil")
	}
}

func TestHandleCallbackStoreFailure(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	querier := sqlcmocks.NewMockQuerier(ctrl)
	querier.EXPECT().CreateIntegration(gomock.Any(), gomock.Any()).
		Return(sqlc.Integration{}, errors.New("db error"))

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

	err := svc.HandleCallback(context.Background(), "some-code")
	if err == nil {
		t.Fatal("expected error for store failure, got nil")
	}
	if !strings.Contains(err.Error(), "could not store integration") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestHandleCallbackSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	querier := sqlcmocks.NewMockQuerier(ctrl)
	querier.EXPECT().CreateIntegration(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params sqlc.CreateIntegrationParams) (sqlc.Integration, error) {
			if params.Provider != "highlevel" {
				t.Fatalf("provider = %s, want highlevel", params.Provider)
			}
			if params.LocationID != "loc_123" {
				t.Fatalf("location_id = %s, want loc_123", params.LocationID)
			}
			if params.AccessToken == "" {
				t.Fatal("access token was not encrypted and stored")
			}
			if params.RefreshToken == "" {
				t.Fatal("refresh token was not encrypted and stored")
			}
			if params.TokenExpiresAt.IsZero() {
				t.Fatal("token expires at was not set")
			}
			return sqlc.Integration{}, nil
		})

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

	err := svc.HandleCallback(context.Background(), "some-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExchangeCodeSendsCorrectForm(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var capturedBody string
	rt := &capturingRoundTripper{
		statusCode: http.StatusOK,
		body:       `{"access_token":"at","refresh_token":"rt","expires_in":3600}`,
		capture:    &capturedBody,
	}

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	svc := newTestService(t, repo, rt)

	token, err := svc.exchangeCode(context.Background(), "auth-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token.AccessToken != "at" {
		t.Fatalf("access_token = %s, want at", token.AccessToken)
	}
	if token.RefreshToken != "rt" {
		t.Fatalf("refresh_token = %s, want rt", token.RefreshToken)
	}

	if !strings.Contains(capturedBody, "grant_type=authorization_code") {
		t.Fatalf("form missing grant_type: %s", capturedBody)
	}
	if !strings.Contains(capturedBody, "client_id=test-client-id") {
		t.Fatalf("form missing client_id: %s", capturedBody)
	}
	if !strings.Contains(capturedBody, "client_secret=test-client-secret") {
		t.Fatalf("form missing client_secret: %s", capturedBody)
	}
	if !strings.Contains(capturedBody, "redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Foauth%2Fcallback") {
		t.Fatalf("form missing redirect_uri: %s", capturedBody)
	}
	if !strings.Contains(capturedBody, "code=auth-code") {
		t.Fatalf("form missing code: %s", capturedBody)
	}
}

func TestExchangeCodeInvalidJSON(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	svc := newTestService(t, repo, &mockRoundTripper{
		statusCode: http.StatusOK,
		body:       `not-json`,
	})

	_, err := svc.exchangeCode(context.Background(), "auth-code")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestEncryptToken(t *testing.T) {
	t.Parallel()

	key := []byte("0123456789abcdef0123456789abcdef")
	plaintext := "super-secret-token"

	encrypted, err := encryptToken(plaintext, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if encrypted == "" {
		t.Fatal("encrypted token is empty")
	}
	if encrypted == plaintext {
		t.Fatal("encrypted token equals plaintext")
	}

	// The output is base64-encoded nonce + ciphertext.
	decoded, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatalf("encrypted token is not valid base64: %v", err)
	}
	if len(decoded) == 0 {
		t.Fatal("decoded encrypted token is empty")
	}
}

func TestEncryptTokenInvalidKey(t *testing.T) {
	t.Parallel()

	// AES-256 requires a 32-byte key.
	_, err := encryptToken("secret", []byte("short"))
	if err == nil {
		t.Fatal("expected error for invalid key length, got nil")
	}
}

type capturingRoundTripper struct {
	statusCode int
	body       string
	capture    *string
}

func (c *capturingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	*c.capture = string(body)

	return &http.Response{
		StatusCode: c.statusCode,
		Body:       io.NopCloser(strings.NewReader(c.body)),
		Header:     make(http.Header),
	}, nil
}
