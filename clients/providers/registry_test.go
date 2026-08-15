package providers

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/rs/zerolog"
)

func TestProviderRegistry(t *testing.T) {
	t.Parallel()

	logger := zerolog.Nop()
	registry := NewProviderRegistry()

	// Test empty registry
	if got := len(registry.List()); got != 0 {
		t.Fatalf("empty registry should have 0 providers, got %d", got)
	}

	// Test registration
	provider := NewHighLevelProvider("client-id", "client-secret", "https://example.com/callback", "webhook-secret", nil)
	registry.Register(provider)

	if got := len(registry.List()); got != 1 {
		t.Fatalf("registry should have 1 provider after registration, got %d", got)
	}

	// Test retrieval
	p, ok := registry.Get("highlevel")
	if !ok {
		t.Fatal("provider should be found by ID")
	}
	if p.ID() != "highlevel" {
		t.Fatalf("provider ID = %s, want highlevel", p.ID())
	}

	// Test unknown provider
	_, ok = registry.Get("unknown")
	if ok {
		t.Fatal("unknown provider should not be found")
	}

	// Test capabilities
	caps := p.Capabilities()
	if len(caps) == 0 {
		t.Fatal("provider should have capabilities")
	}

	if !p.HasCapability(CapabilityOAuth) {
		t.Fatal("provider should have OAuth capability")
	}

	if !p.HasCapability(CapabilityWebhooks) {
		t.Fatal("provider should have Webhooks capability")
	}

	if p.HasCapability(CapabilityHealthCheck) {
		t.Fatal("provider should not have HealthCheck capability")
	}

	// Test GetByCapability
	oauthProviders := registry.GetByCapability(CapabilityOAuth)
	if len(oauthProviders) != 1 {
		t.Fatalf("should find 1 OAuth provider, got %d", len(oauthProviders))
	}

	webhookProviders := registry.GetByCapability(CapabilityWebhooks)
	if len(webhookProviders) != 1 {
		t.Fatalf("should find 1 Webhook provider, got %d", len(webhookProviders))
	}

	healthProviders := registry.GetByCapability(CapabilityHealthCheck)
	if len(healthProviders) != 0 {
		t.Fatalf("should find 0 HealthCheck providers, got %d", len(healthProviders))
	}

	// Test duplicate registration (should overwrite)
	provider2 := NewHighLevelProvider("client-id-2", "client-secret-2", "https://example.com/callback2", "webhook-secret-2", nil)
	registry.Register(provider2)
	if got := len(registry.List()); got != 1 {
		t.Fatalf("registry should still have 1 provider after duplicate registration, got %d", got)
	}

	logger.Info().Msg("provider registry tests passed")
}

func TestHighLevelWebhookPublicKeyIsUsedForSignatureVerification(t *testing.T) {
	t.Parallel()

	// SECURITY REGRESSION TEST (SEC-03): the webhook signature must be verified
	// with the distinct Ed25519 public key (HIGHLEVEL_WEBHOOK_PUBLIC_KEY), never
	// the OAuth client secret. A signature produced by the matching private key
	// must verify; a signature from a different key must be rejected.
	clientSecret := "oauth-client-secret"

	// Generate a key pair for the webhook public key.
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate ed25519 key: %v", err)
	}
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	webhookPublicKey := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))

	provider := NewHighLevelProvider("client-id", clientSecret, "https://example.com/callback", webhookPublicKey, nil)

	whp := provider.WebhookProvider()
	if whp == nil {
		t.Fatal("webhook provider should not be nil")
	}

	body := []byte(`{"eventId":"evt_1","eventType":"integration.installed"}`)

	sign := func(key ed25519.PrivateKey) string {
		sig := ed25519.Sign(key, body)
		return base64.StdEncoding.EncodeToString(sig)
	}

	headers := map[string]string{
		"X-GHL-Signature": sign(priv),
	}
	if err := whp.VerifyRequest(context.Background(), headers, body); err != nil {
		t.Fatalf("signature with the matching private key must verify, got error: %v", err)
	}

	// A signature from a different key must be rejected.
	_, otherPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate other ed25519 key: %v", err)
	}
	badHeaders := map[string]string{
		"X-GHL-Signature": sign(otherPriv),
	}
	if err := whp.VerifyRequest(context.Background(), badHeaders, body); err == nil {
		t.Fatal("signature from a different key must be rejected")
	}
}
