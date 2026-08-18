package providers

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/rs/zerolog"
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
		Type       string                 `json:"type"`
		AppID      string                 `json:"appId"`
		LocationID string                 `json:"locationId"`
		CompanyID  string                 `json:"companyId"`
		Data       map[string]interface{} `json:"data"`
		Timestamp  string                 `json:"timestamp"`
		WebhookID  string                 `json:"webhookId"`
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
//
// It resolves GHL webhook events to the specific RVPay integration and client
// using the deterministic GHL locationId mapping. The INSTALL event creates or
// finds the payment_provider_configs record idempotently so that the OAuth
// callback and INSTALL webhook can arrive in any order.
type HighLevelWebhookDispatcher struct {
	logger           Logger
	integrationsRepo repo.IntegrationRepo
	configRepo       repo.PaymentProviderConfigRepo
	providerConfig   ProviderConfigSettings
}

// ProviderConfigSettings holds the configuration used to build the HighLevel
// Custom Payment Provider configuration. All values come from environment
// configuration; none are hard-coded.
type ProviderConfigSettings struct {
	// Name is the display name of the payment provider.
	Name string
	// Description is the description of the payment provider.
	Description string
	// ImageURL is the publicly accessible image URL of the payment provider.
	ImageURL string
	// PaymentsURL is the frontend checkout URL supplied to HighLevel.
	PaymentsURL string
	// QueryURL is the backend query URL supplied to HighLevel.
	QueryURL string
}

// NewHighLevelWebhookDispatcher creates a new HighLevel webhook dispatcher.
//
// integrationsRepo and configRepo are required to resolve the GHL locationId
// to the specific RVPay integration and to create/find the
// payment_provider_configs record idempotently. providerConfig holds the
// configuration used when creating a new payment provider config from an
// INSTALL event.
func NewHighLevelWebhookDispatcher(
	logger Logger,
	integrationsRepo repo.IntegrationRepo,
	configRepo repo.PaymentProviderConfigRepo,
	providerConfig ProviderConfigSettings,
) *HighLevelWebhookDispatcher {
	return &HighLevelWebhookDispatcher{
		logger:           logger,
		integrationsRepo: integrationsRepo,
		configRepo:       configRepo,
		providerConfig:   providerConfig,
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

// handleIntegrationInstalled processes a GHL INSTALL event. It resolves the
// specific RVPay integration from the GHL locationId and creates or finds the
// corresponding payment_provider_configs record idempotently.
//
// Resolution order:
//  1. integration.external_account_id = GHL locationId (the deterministic
//     provisioning mapping established when the integration is activated).
//  2. payment_provider_configs.location_id = GHL locationId (created during
//     provider registration).
//
// If the integration cannot be resolved, the event fails clearly rather than
// selecting an arbitrary integration or client. This preserves multiple-client
// safety: a GHL installation belonging to Client B never resolves to Client A.
func (d *HighLevelWebhookDispatcher) handleIntegrationInstalled(ctx context.Context, event *WebhookEvent) error {
	if event.LocationID == "" {
		return fmt.Errorf("INSTALL event missing locationId")
	}
	if d.configRepo == nil {
		return fmt.Errorf("payment provider config repo not configured")
	}
	if d.integrationsRepo == nil {
		return fmt.Errorf("integration repo not configured")
	}

	// Resolve the integration deterministically from the GHL locationId.
	integration, err := d.integrationsRepo.GetByExternalAccountID(ctx, event.LocationID)
	if err == repo.ErrNotFound {
		// Fall back to the payment provider config mapping.
		config, configErr := d.configRepo.GetByLocationID(ctx, event.LocationID)
		if configErr == repo.ErrNotFound {
			return fmt.Errorf("no integration found for GHL locationId %q", event.LocationID)
		}
		if configErr != nil {
			return fmt.Errorf("resolve payment provider config for locationId: %w", configErr)
		}
		integration, err = d.integrationsRepo.GetByID(ctx, config.IntegrationID)
		if err == repo.ErrNotFound {
			return fmt.Errorf("integration %s not found for GHL locationId %q", config.IntegrationID, event.LocationID)
		}
		if err != nil {
			return fmt.Errorf("resolve integration for locationId: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("resolve integration by external account id: %w", err)
	}

	// Create or find the payment_provider_configs record idempotently.
	// If the config already exists for this integration, reuse it. Otherwise
	// create it with the GHL locationId. The provider name/description/URLs
	// are populated from configuration; they may be empty if not configured,
	// and will be completed during provider registration.
	_, err = d.configRepo.GetByIntegrationID(ctx, integration.ID)
	if err == nil {
		d.logger.Info("INSTALL event: payment provider config already exists; reusing", "integration_id", integration.ID.String(), "location_id", event.LocationID)
		return nil
	}
	if !errors.Is(err, repo.ErrNotFound) {
		return fmt.Errorf("get payment provider config for integration: %w", err)
	}

	_, err = d.configRepo.Create(
		ctx,
		integration.ID,
		d.providerConfig.Name,
		d.providerConfig.Description,
		d.providerConfig.ImageURL,
		event.LocationID,
		d.providerConfig.QueryURL,
		d.providerConfig.PaymentsURL,
		false, // RVPay supports one-time payments only.
		"",    // provider API key is generated during provider registration.
	)
	if err == repo.ErrDuplicate {
		// A concurrent INSTALL event created the config; reuse it.
		d.logger.Info("INSTALL event: payment provider config created concurrently; reusing", "integration_id", integration.ID.String(), "location_id", event.LocationID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("create payment provider config for integration: %w", err)
	}

	d.logger.Info("INSTALL event: payment provider config created", "integration_id", integration.ID.String(), "location_id", event.LocationID)
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

// highLevelWebhookLogger adapts a zerolog.Logger to the providers.Logger
// interface used by the HighLevelWebhookDispatcher.
type highLevelWebhookLogger struct {
	logger zerolog.Logger
}

// NewHighLevelWebhookLogger creates a Logger adapter backed by a zerolog
// logger. It is used to wire the HighLevel webhook dispatcher with the
// service's structured logger.
func NewHighLevelWebhookLogger(logger zerolog.Logger) Logger {
	return &highLevelWebhookLogger{logger: logger}
}

func (l *highLevelWebhookLogger) Info(msg string, args ...interface{}) {
	event := l.logger.Info()
	for i := 0; i+1 < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		event = event.Interface(key, args[i+1])
	}
	event.Msg(msg)
}
