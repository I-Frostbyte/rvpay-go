package providers

import (
	"context"
)

// WebhookEvent represents a normalized webhook event from any provider.
type WebhookEvent struct {
	Provider        string
	EventType       string
	ProviderEventID string
	IntegrationID   string
	ClientID        string
	PlatformID      string
	LocationID      string
	Payload         map[string]interface{}
	ReceivedAt      int64
}

// WebhookProvider defines webhook operations for a provider.
type WebhookProvider interface {
	// VerifyRequest validates the incoming webhook request.
	VerifyRequest(ctx context.Context, headers map[string]string, body []byte) error
	// ParseEvent converts a provider payload into a normalized WebhookEvent.
	ParseEvent(ctx context.Context, body []byte) (*WebhookEvent, error)
	// RegisterWebhook registers a webhook subscription with the provider.
	RegisterWebhook(ctx context.Context, integrationID, callbackURL string) error
	// UnregisterWebhook removes a webhook subscription from the provider.
	UnregisterWebhook(ctx context.Context, integrationID string) error
}

// WebhookValidator defines signature validation for webhooks.
type WebhookValidator interface {
	// ValidateSignature verifies the webhook signature.
	ValidateSignature(secret string, headers map[string]string, body []byte) error
}

// WebhookDispatcher defines event dispatching to business services.
type WebhookDispatcher interface {
	// Dispatch sends a normalized event to the appropriate handler.
	Dispatch(ctx context.Context, event *WebhookEvent) error
}
