package providers

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"
	"time"
)

// HighLevelWebhookProvider implements WebhookProvider for HighLevel.
//
// HighLevel signs webhook deliveries with an Ed25519 signature carried in the
// X-GHL-Signature header. The signature is computed over the exact raw request
// body bytes, so verification MUST operate on the original body and never on a
// re-marshaled or transformed JSON representation.
type HighLevelWebhookProvider struct {
	publicKey ed25519.PublicKey
}

// NewHighLevelWebhookProvider creates a new HighLevel webhook provider.
// publicKeyPEM is the PEM-encoded Ed25519 public key (HIGHLEVEL_WEBHOOK_PUBLIC_KEY).
func NewHighLevelWebhookProvider(publicKeyPEM string) *HighLevelWebhookProvider {
	return &HighLevelWebhookProvider{
		publicKey: parseEd25519PublicKey(publicKeyPEM),
	}
}

// VerifyRequest validates the X-GHL-Signature header against the raw body
// using the configured Ed25519 public key. It rejects missing, malformed, and
// invalid signatures. The body passed in MUST be the exact raw request body.
func (p *HighLevelWebhookProvider) VerifyRequest(ctx context.Context, headers map[string]string, body []byte) error {
	if p.publicKey == nil {
		return fmt.Errorf("webhook public key is not configured")
	}

	signature := headers["X-GHL-Signature"]
	if signature == "" {
		return fmt.Errorf("missing signature header")
	}

	// GHL encodes the Ed25519 signature as base64 (standard alphabet).
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("malformed signature: %w", err)
	}

	if len(sig) != ed25519.SignatureSize {
		return fmt.Errorf("malformed signature: unexpected length %d", len(sig))
	}

	if !ed25519.Verify(p.publicKey, body, sig) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// ParseEvent converts a provider payload into a normalized WebhookEvent.
//
// HighLevel webhook payloads use the GHL schema:
//
//	{
//	  "type": "INSTALL",
//	  "appId": "...",
//	  "locationId": "...",
//	  "companyId": "...",
//	  "timestamp": "2026-08-17T09:06:59.366Z",
//	  "webhookId": "..."
//	}
//
// Mapping:
//   - type → EventType
//   - webhookId → ProviderEventID
//   - appId → IntegrationID (GHL Marketplace app ID, NOT an RVPay UUID)
//   - companyId → ClientID
//   - locationId → LocationID
func (p *HighLevelWebhookProvider) ParseEvent(ctx context.Context, body []byte) (*WebhookEvent, error) {
	var payload struct {
		Type         string                 `json:"type"`
		AppID        string                 `json:"appId"`
		LocationID   string                 `json:"locationId"`
		CompanyID    string                 `json:"companyId"`
		Data         map[string]interface{} `json:"data"`
		Timestamp    string                 `json:"timestamp"`
		WebhookID    string                 `json:"webhookId"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	timestamp, err := time.Parse(time.RFC3339Nano, payload.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid webhook timestamp: %w", err)
	}

	event := &WebhookEvent{
		Provider:        "highlevel",
		EventType:       payload.Type,
		ProviderEventID: payload.WebhookID,
		IntegrationID:   payload.AppID,
		ClientID:        payload.CompanyID,
		LocationID:      payload.LocationID,
		Payload:         payload.Data,
		ReceivedAt:      timestamp.Unix(),
	}

	return event, nil
}

func (p *HighLevelWebhookProvider) RegisterWebhook(ctx context.Context, integrationID, callbackURL string) error {
	// In a real implementation, this would call the HighLevel API to register the webhook
	// For now, this is a stub that would be implemented with actual API calls
	return nil
}

func (p *HighLevelWebhookProvider) UnregisterWebhook(ctx context.Context, integrationID string) error {
	// In a real implementation, this would call the HighLevel API to unregister the webhook
	// For now, this is a stub that would be implemented with actual API calls
	return nil
}

// parseEd25519PublicKey parses a PEM-encoded Ed25519 public key. It returns
// nil when the key is empty or malformed so that verification fails closed.
func parseEd25519PublicKey(publicKeyPEM string) ed25519.PublicKey {
	trimmed := strings.TrimSpace(publicKeyPEM)
	if trimmed == "" {
		return nil
	}

	block, _ := pem.Decode([]byte(trimmed))
	if block == nil {
		return nil
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil
	}

	edKey, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil
	}

	return edKey
}

// HighLevelWebhookValidator implements WebhookValidator for HighLevel.
type HighLevelWebhookValidator struct {
	publicKey ed25519.PublicKey
}

// NewHighLevelWebhookValidator creates a new HighLevel webhook validator.
func NewHighLevelWebhookValidator(publicKeyPEM string) *HighLevelWebhookValidator {
	return &HighLevelWebhookValidator{
		publicKey: parseEd25519PublicKey(publicKeyPEM),
	}
}

func (v *HighLevelWebhookValidator) ValidateSignature(secret string, headers map[string]string, body []byte) error {
	if v.publicKey == nil {
		return fmt.Errorf("webhook public key is not configured")
	}

	signature := headers["X-GHL-Signature"]
	if signature == "" {
		return fmt.Errorf("missing signature header")
	}

	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("malformed signature: %w", err)
	}

	if len(sig) != ed25519.SignatureSize {
		return fmt.Errorf("malformed signature: unexpected length %d", len(sig))
	}

	if !ed25519.Verify(v.publicKey, body, sig) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// HighLevelWebhookDispatcher implements WebhookDispatcher for HighLevel.
type HighLevelWebhookDispatcher struct {
	logger Logger
}

// NewHighLevelWebhookDispatcher creates a new HighLevel webhook dispatcher.
func NewHighLevelWebhookDispatcher(logger Logger) *HighLevelWebhookDispatcher {
	return &HighLevelWebhookDispatcher{
		logger: logger,
	}
}

func (d *HighLevelWebhookDispatcher) Dispatch(ctx context.Context, event *WebhookEvent) error {
	d.logger.Info("Dispatching webhook event", "provider", event.Provider, "event_type", event.EventType, "event_id", event.ProviderEventID)

	switch event.EventType {
	case "INSTALL", "integration.installed":
		return d.handleIntegrationInstalled(ctx, event)
	case "UNINSTALL", "integration.uninstalled":
		return d.handleIntegrationUninstalled(ctx, event)
	case "oauth.revoked":
		return d.handleOAuthRevoked(ctx, event)
	case "token.expired":
		return d.handleTokenExpired(ctx, event)
	case "provider.disconnected":
		return d.handleProviderDisconnected(ctx, event)
	default:
		d.logger.Info("Unhandled webhook event type", "event_type", event.EventType)
		return nil
	}
}

func (d *HighLevelWebhookDispatcher) handleIntegrationInstalled(ctx context.Context, event *WebhookEvent) error {
	d.logger.Info("Integration installed event received", "integration_id", event.IntegrationID)
	// Dispatch to integrations service to update integration status
	return nil
}

func (d *HighLevelWebhookDispatcher) handleIntegrationUninstalled(ctx context.Context, event *WebhookEvent) error {
	d.logger.Info("Integration uninstalled event received", "integration_id", event.IntegrationID)
	// Dispatch to integrations service to remove integration
	return nil
}

func (d *HighLevelWebhookDispatcher) handleOAuthRevoked(ctx context.Context, event *WebhookEvent) error {
	d.logger.Info("OAuth revoked event received", "integration_id", event.IntegrationID)
	// Dispatch to OAuth service to invalidate tokens
	return nil
}

func (d *HighLevelWebhookDispatcher) handleTokenExpired(ctx context.Context, event *WebhookEvent) error {
	d.logger.Info("Token expired event received", "integration_id", event.IntegrationID)
	// Dispatch to OAuth service to refresh tokens
	return nil
}

func (d *HighLevelWebhookDispatcher) handleProviderDisconnected(ctx context.Context, event *WebhookEvent) error {
	d.logger.Info("Provider disconnected event received", "integration_id", event.IntegrationID)
	// Dispatch to integrations service to mark integration as disconnected
	return nil
}

// Logger defines the logging interface.
type Logger interface {
	Info(msg string, args ...interface{})
}
