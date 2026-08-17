package providers

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"
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
