package providers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// HighLevelWebhookProvider implements WebhookProvider for HighLevel.
type HighLevelWebhookProvider struct {
	webhookSecret string
}

// NewHighLevelWebhookProvider creates a new HighLevel webhook provider.
func NewHighLevelWebhookProvider(webhookSecret string) *HighLevelWebhookProvider {
	return &HighLevelWebhookProvider{
		webhookSecret: webhookSecret,
	}
}

func (p *HighLevelWebhookProvider) VerifyRequest(ctx context.Context, headers map[string]string, body []byte) error {
	signature := headers["X-HighLevel-Signature"]
	if signature == "" {
		return fmt.Errorf("missing signature header")
	}

	timestamp := headers["X-HighLevel-Timestamp"]
	if timestamp == "" {
		return fmt.Errorf("missing timestamp header")
	}

	expectedSignature := p.computeSignature(timestamp, body)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func (p *HighLevelWebhookProvider) ParseEvent(ctx context.Context, body []byte) (*WebhookEvent, error) {
	var payload struct {
		EventID       string                 `json:"eventId"`
		EventType     string                 `json:"eventType"`
		IntegrationID string                 `json:"integrationId"`
		ClientID      string                 `json:"clientId"`
		Data          map[string]interface{} `json:"data"`
		Timestamp     int64                  `json:"timestamp"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	event := &WebhookEvent{
		Provider:        "highlevel",
		EventType:       payload.EventType,
		ProviderEventID: payload.EventID,
		IntegrationID:   payload.IntegrationID,
		ClientID:        payload.ClientID,
		Payload:         payload.Data,
		ReceivedAt:      payload.Timestamp,
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

func (p *HighLevelWebhookProvider) computeSignature(timestamp string, body []byte) string {
	data := timestamp + string(body)
	h := hmac.New(sha256.New, []byte(p.webhookSecret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// HighLevelWebhookValidator implements WebhookValidator for HighLevel.
type HighLevelWebhookValidator struct {
	secret string
}

// NewHighLevelWebhookValidator creates a new HighLevel webhook validator.
func NewHighLevelWebhookValidator(secret string) *HighLevelWebhookValidator {
	return &HighLevelWebhookValidator{
		secret: secret,
	}
}

func (v *HighLevelWebhookValidator) ValidateSignature(secret string, headers map[string]string, body []byte) error {
	signature := headers["X-HighLevel-Signature"]
	if signature == "" {
		return fmt.Errorf("missing signature header")
	}

	timestamp := headers["X-HighLevel-Timestamp"]
	if timestamp == "" {
		return fmt.Errorf("missing timestamp header")
	}

	expectedSignature := v.computeSignature(timestamp, body)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func (v *HighLevelWebhookValidator) computeSignature(timestamp string, body []byte) string {
	data := timestamp + string(body)
	h := hmac.New(sha256.New, []byte(v.secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
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
	case "integration.installed":
		return d.handleIntegrationInstalled(ctx, event)
	case "integration.uninstalled":
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
