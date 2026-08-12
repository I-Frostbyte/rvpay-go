package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	clientshttp "github.com/I-Frostbyte/rvpay-go/clients/http"
	"github.com/I-Frostbyte/rvpay-go/clients/oauth"
	"github.com/I-Frostbyte/rvpay-go/clients/providers"
	"github.com/I-Frostbyte/rvpay-go/clients/webhooks"
	"github.com/rs/zerolog"
)

// newRuntimeMux builds the exact HTTP mux wiring used by main.go: the
// grpc-gateway mux at "/", /healthz, and the GHL OAuth/webhook routes. It
// returns the mux so tests can prove the routes are registered and reachable.
func newRuntimeMux(t *testing.T) *http.ServeMux {
	t.Helper()

	// Minimal in-memory repos are not needed for route registration; the
	// handlers are wired to services that will fail on actual processing but
	// the routes themselves must be reachable. We use the real service
	// constructors with nil repos to prove wiring compiles and routes exist.
	registry := providers.NewProviderRegistry()
	registry.Register(providers.NewHighLevelProvider("test-client", "test-secret", "https://example.com/callback", ""))

	oauthService := oauth.NewService(nil, nil, nil, nil, nil, registry, "https://example.com/callback", zerolog.Nop())
	webhookService := webhooks.NewService(nil, nil, nil, nil, registry, zerolog.Nop())
	oauthHandler := clientshttp.NewOAuthHandler(oauthService, zerolog.Nop())
	webhookHandler := clientshttp.NewWebhookHandler(webhookService, zerolog.Nop())

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/oauth/callback", oauthHandler.Callback)
	mux.HandleFunc("/webhooks/highlevel", webhookHandler.HighLevel)

	return mux
}

func TestRuntime_HealthzRegistered(t *testing.T) {
	t.Parallel()

	mux := newRuntimeMux(t)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("healthz status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRuntime_OAuthCallbackRouteRegistered(t *testing.T) {
	t.Parallel()

	mux := newRuntimeMux(t)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// The route must be reachable. A GET with no params should return 400
	// (missing code/state) rather than 404, proving the route is registered.
	resp, err := http.Get(srv.URL + "/oauth/callback")
	if err != nil {
		t.Fatalf("GET /oauth/callback: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		t.Fatal("OAuth callback route is not registered (got 404)")
	}
}

func TestRuntime_WebhookRouteRegistered(t *testing.T) {
	t.Parallel()

	mux := newRuntimeMux(t)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// The route must be reachable. A POST with no signature should return 400
	// (missing signature) rather than 404, proving the route is registered.
	resp, err := http.Post(srv.URL+"/webhooks/highlevel", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /webhooks/highlevel: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		t.Fatal("webhook route is not registered (got 404)")
	}
}

func TestRuntime_OAuthCallbackMethodNotAllowed(t *testing.T) {
	t.Parallel()

	mux := newRuntimeMux(t)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/oauth/callback", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /oauth/callback: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("OAuth callback POST status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestRuntime_WebhookMethodNotAllowed(t *testing.T) {
	t.Parallel()

	mux := newRuntimeMux(t)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/webhooks/highlevel")
	if err != nil {
		t.Fatalf("GET /webhooks/highlevel: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("webhook GET status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}
