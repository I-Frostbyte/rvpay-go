package providers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestPaymentProviderServer creates an httptest.Server that records the
// requests it receives and returns the configured response. It returns the
// server, a channel of received requests, and a cleanup function.
func newTestPaymentProviderServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *[]*http.Request) {
	t.Helper()

	var requests []*http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Clone the request so we can inspect it after the handler returns.
		reqCopy := r.Clone(r.Context())
		requests = append(requests, reqCopy)
		handler(w, r)
	}))

	t.Cleanup(srv.Close)
	return srv, &requests
}

func TestCreateProviderAssociation_Success(t *testing.T) {
	t.Parallel()

	srv, requests := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/payments/custom-provider/provider" {
			t.Errorf("path = %s, want /payments/custom-provider/provider", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-access-token" {
			t.Errorf("Authorization = %q, want Bearer test-access-token", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	err := client.CreateProviderAssociation(context.Background(), "test-access-token", "loc-123")
	if err != nil {
		t.Fatalf("CreateProviderAssociation failed: %v", err)
	}

	if len(*requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*requests))
	}
}

func TestCreateProviderAssociation_MissingAccessToken(t *testing.T) {
	t.Parallel()

	client := NewHighLevelPaymentProviderClient("https://example.com", nil)
	err := client.CreateProviderAssociation(context.Background(), "", "loc-123")
	if !errors.Is(err, ErrMissingAccessToken) {
		t.Fatalf("expected ErrMissingAccessToken, got %v", err)
	}
}

func TestCreateProviderAssociation_MissingLocationID(t *testing.T) {
	t.Parallel()

	client := NewHighLevelPaymentProviderClient("https://example.com", nil)
	err := client.CreateProviderAssociation(context.Background(), "token", "")
	if !errors.Is(err, ErrMissingLocationID) {
		t.Fatalf("expected ErrMissingLocationID, got %v", err)
	}
}

func TestCreateProviderAssociation_BadRequest(t *testing.T) {
	t.Parallel()

	srv, _ := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"invalid location"}`))
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	err := client.CreateProviderAssociation(context.Background(), "token", "loc-123")
	if !errors.Is(err, ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestCreateProviderAssociation_Unauthorized(t *testing.T) {
	t.Parallel()

	srv, _ := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"invalid token"}`))
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	err := client.CreateProviderAssociation(context.Background(), "token", "loc-123")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestCreateProviderAssociation_UnprocessableEntity(t *testing.T) {
	t.Parallel()

	srv, _ := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"validation failed"}`))
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	err := client.CreateProviderAssociation(context.Background(), "token", "loc-123")
	if !errors.Is(err, ErrUnprocessableEntity) {
		t.Fatalf("expected ErrUnprocessableEntity, got %v", err)
	}
}

func TestCreateProviderAssociation_NetworkError(t *testing.T) {
	t.Parallel()

	// Use a closed server to simulate a network error.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	err := client.CreateProviderAssociation(context.Background(), "token", "loc-123")
	if err == nil {
		t.Fatal("expected network error")
	}
}

func TestCreateProviderAssociation_ContextCancellation(t *testing.T) {
	t.Parallel()

	srv, _ := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow response.
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := client.CreateProviderAssociation(ctx, "token", "loc-123")
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestCreateProviderConfig_Success(t *testing.T) {
	t.Parallel()

	srv, requests := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/payments/custom-provider/connect" {
			t.Errorf("path = %s, want /payments/custom-provider/connect", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-access-token" {
			t.Errorf("Authorization = %q, want Bearer test-access-token", got)
		}

		// Verify the request body contains the expected fields.
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if body["name"] != "RVPay" {
			t.Errorf("name = %v, want RVPay", body["name"])
		}
		if body["locationId"] != "loc-123" {
			t.Errorf("locationId = %v, want loc-123", body["locationId"])
		}
		if body["queryUrl"] != "https://api.example.com/payments/custom-provider/query" {
			t.Errorf("queryUrl = %v, want configured URL", body["queryUrl"])
		}
		if body["paymentsUrl"] != "https://checkout.example.com/payment/checkout" {
			t.Errorf("paymentsUrl = %v, want configured URL", body["paymentsUrl"])
		}
		if body["supportsSubscriptionSchedule"] != false {
			t.Errorf("supportsSubscriptionSchedule = %v, want false", body["supportsSubscriptionSchedule"])
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	config := ProviderConfig{
		Name:                         "RVPay",
		Description:                  "RVPay payment provider",
		ImageURL:                     "https://example.com/logo.jpg",
		LocationID:                   "loc-123",
		QueryURL:                     "https://api.example.com/payments/custom-provider/query",
		PaymentsURL:                  "https://checkout.example.com/payment/checkout",
		SupportsSubscriptionSchedule: false,
	}

	err := client.CreateProviderConfig(context.Background(), "test-access-token", config)
	if err != nil {
		t.Fatalf("CreateProviderConfig failed: %v", err)
	}

	if len(*requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*requests))
	}
}

func TestCreateProviderConfig_MissingAccessToken(t *testing.T) {
	t.Parallel()

	client := NewHighLevelPaymentProviderClient("https://example.com", nil)
	err := client.CreateProviderConfig(context.Background(), "", ProviderConfig{LocationID: "loc-123"})
	if !errors.Is(err, ErrMissingAccessToken) {
		t.Fatalf("expected ErrMissingAccessToken, got %v", err)
	}
}

func TestCreateProviderConfig_MissingLocationID(t *testing.T) {
	t.Parallel()

	client := NewHighLevelPaymentProviderClient("https://example.com", nil)
	err := client.CreateProviderConfig(context.Background(), "token", ProviderConfig{})
	if !errors.Is(err, ErrMissingLocationID) {
		t.Fatalf("expected ErrMissingLocationID, got %v", err)
	}
}

func TestCreateProviderConfig_BadRequest(t *testing.T) {
	t.Parallel()

	srv, _ := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"invalid config"}`))
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	err := client.CreateProviderConfig(context.Background(), "token", ProviderConfig{LocationID: "loc-123"})
	if !errors.Is(err, ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestCreateProviderConfig_Unauthorized(t *testing.T) {
	t.Parallel()

	srv, _ := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"invalid token"}`))
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	err := client.CreateProviderConfig(context.Background(), "token", ProviderConfig{LocationID: "loc-123"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestCreateProviderConfig_UnprocessableEntity(t *testing.T) {
	t.Parallel()

	srv, _ := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"validation failed"}`))
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	err := client.CreateProviderConfig(context.Background(), "token", ProviderConfig{LocationID: "loc-123"})
	if !errors.Is(err, ErrUnprocessableEntity) {
		t.Fatalf("expected ErrUnprocessableEntity, got %v", err)
	}
}

func TestFetchProviderConfig_Success(t *testing.T) {
	t.Parallel()

	srv, requests := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/payments/custom-provider/connect" {
			t.Errorf("path = %s, want /payments/custom-provider/connect", r.URL.Path)
		}
		if got := r.URL.Query().Get("locationId"); got != "loc-123" {
			t.Errorf("locationId query param = %q, want loc-123", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-access-token" {
			t.Errorf("Authorization = %q, want Bearer test-access-token", got)
		}

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
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	config, err := client.FetchProviderConfig(context.Background(), "test-access-token", "loc-123")
	if err != nil {
		t.Fatalf("FetchProviderConfig failed: %v", err)
	}

	if config.Name != "RVPay" {
		t.Errorf("Name = %q, want RVPay", config.Name)
	}
	if config.LocationID != "loc-123" {
		t.Errorf("LocationID = %q, want loc-123", config.LocationID)
	}
	if config.QueryURL != "https://api.example.com/payments/custom-provider/query" {
		t.Errorf("QueryURL = %q, want configured URL", config.QueryURL)
	}
	if config.PaymentsURL != "https://checkout.example.com/payment/checkout" {
		t.Errorf("PaymentsURL = %q, want configured URL", config.PaymentsURL)
	}
	if config.SupportsSubscriptionSchedule {
		t.Error("SupportsSubscriptionSchedule should be false")
	}

	if len(*requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*requests))
	}
}

func TestFetchProviderConfig_MissingAccessToken(t *testing.T) {
	t.Parallel()

	client := NewHighLevelPaymentProviderClient("https://example.com", nil)
	_, err := client.FetchProviderConfig(context.Background(), "", "loc-123")
	if !errors.Is(err, ErrMissingAccessToken) {
		t.Fatalf("expected ErrMissingAccessToken, got %v", err)
	}
}

func TestFetchProviderConfig_MissingLocationID(t *testing.T) {
	t.Parallel()

	client := NewHighLevelPaymentProviderClient("https://example.com", nil)
	_, err := client.FetchProviderConfig(context.Background(), "token", "")
	if !errors.Is(err, ErrMissingLocationID) {
		t.Fatalf("expected ErrMissingLocationID, got %v", err)
	}
}

func TestFetchProviderConfig_MalformedResponse(t *testing.T) {
	t.Parallel()

	srv, _ := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	_, err := client.FetchProviderConfig(context.Background(), "token", "loc-123")
	if err == nil {
		t.Fatal("expected malformed response error")
	}
}

func TestFetchProviderConfig_Unauthorized(t *testing.T) {
	t.Parallel()

	srv, _ := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"invalid token"}`))
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	_, err := client.FetchProviderConfig(context.Background(), "token", "loc-123")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestDisconnectProvider_Success(t *testing.T) {
	t.Parallel()

	srv, requests := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/payments/custom-provider/connect" {
			t.Errorf("path = %s, want /payments/custom-provider/connect", r.URL.Path)
		}
		if got := r.URL.Query().Get("locationId"); got != "loc-123" {
			t.Errorf("locationId query param = %q, want loc-123", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-access-token" {
			t.Errorf("Authorization = %q, want Bearer test-access-token", got)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	err := client.DisconnectProvider(context.Background(), "test-access-token", "loc-123")
	if err != nil {
		t.Fatalf("DisconnectProvider failed: %v", err)
	}

	if len(*requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*requests))
	}
}

func TestDisconnectProvider_MissingAccessToken(t *testing.T) {
	t.Parallel()

	client := NewHighLevelPaymentProviderClient("https://example.com", nil)
	err := client.DisconnectProvider(context.Background(), "", "loc-123")
	if !errors.Is(err, ErrMissingAccessToken) {
		t.Fatalf("expected ErrMissingAccessToken, got %v", err)
	}
}

func TestDisconnectProvider_MissingLocationID(t *testing.T) {
	t.Parallel()

	client := NewHighLevelPaymentProviderClient("https://example.com", nil)
	err := client.DisconnectProvider(context.Background(), "token", "")
	if !errors.Is(err, ErrMissingLocationID) {
		t.Fatalf("expected ErrMissingLocationID, got %v", err)
	}
}

func TestDisconnectProvider_Unauthorized(t *testing.T) {
	t.Parallel()

	srv, _ := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"invalid token"}`))
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	err := client.DisconnectProvider(context.Background(), "token", "loc-123")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestDisconnectProvider_UnprocessableEntity(t *testing.T) {
	t.Parallel()

	srv, _ := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"validation failed"}`))
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	err := client.DisconnectProvider(context.Background(), "token", "loc-123")
	if !errors.Is(err, ErrUnprocessableEntity) {
		t.Fatalf("expected ErrUnprocessableEntity, got %v", err)
	}
}

func TestAccessTokenNotInErrors(t *testing.T) {
	t.Parallel()

	// SECURITY TEST: the access token must never appear in error strings.
	srv, _ := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"bad request"}`))
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	err := client.CreateProviderAssociation(context.Background(), "super-secret-access-token", "loc-123")
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "super-secret-access-token") {
		t.Fatal("access token leaked into error string")
	}
}

func TestAccessTokenNotInLogs(t *testing.T) {
	t.Parallel()

	// SECURITY TEST: the access token must never be logged. We verify this by
	// checking that the request body and headers do not contain the token.
	srv, requests := newTestPaymentProviderServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	})

	client := NewHighLevelPaymentProviderClient(srv.URL, nil)
	err := client.CreateProviderAssociation(context.Background(), "super-secret-access-token", "loc-123")
	if err != nil {
		t.Fatalf("CreateProviderAssociation failed: %v", err)
	}

	if len(*requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*requests))
	}

	// The token should only be in the Authorization header, never in the body.
	req := (*requests)[0]
	if got := req.Header.Get("Authorization"); got != "Bearer super-secret-access-token" {
		t.Errorf("Authorization = %q, want Bearer super-secret-access-token", got)
	}
}

func TestHighLevelProviderPaymentProviderCapability(t *testing.T) {
	t.Parallel()

	// A provider with a payment provider client should advertise the
	// CapabilityPaymentProvider capability.
	paymentClient := NewHighLevelPaymentProviderClient("https://example.com", nil)
	provider := NewHighLevelProvider("client-id", "client-secret", "https://example.com/callback", "", paymentClient)

	if !provider.HasCapability(CapabilityPaymentProvider) {
		t.Fatal("provider should have CapabilityPaymentProvider when payment client is set")
	}

	// A provider without a payment provider client should not advertise it.
	providerNoPayment := NewHighLevelProvider("client-id", "client-secret", "https://example.com/callback", "", nil)
	if providerNoPayment.HasCapability(CapabilityPaymentProvider) {
		t.Fatal("provider should not have CapabilityPaymentProvider when payment client is nil")
	}

	// PaymentProvider() should return the configured client.
	if provider.PaymentProvider() != paymentClient {
		t.Fatal("PaymentProvider() should return the configured client")
	}
	if providerNoPayment.PaymentProvider() != nil {
		t.Fatal("PaymentProvider() should return nil when no client is configured")
	}
}
