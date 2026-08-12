package repo

import (
	"context"

	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/google/uuid"
)

// WebhookEventRepo provides persistence operations for webhook events.
type WebhookEventRepo interface {
	// Create inserts a webhook event. It returns ErrDuplicate when the
	// (integration_id, provider_event_id) pair already exists, enabling
	// race-safe idempotent processing of duplicate deliveries.
	Create(ctx context.Context, integrationID uuid.UUID, providerEventID, eventType string, payload []byte) (sqlc.WebhookEvent, error)
	GetByIntegrationAndProvider(ctx context.Context, integrationID uuid.UUID, providerEventID string) (sqlc.WebhookEvent, error)
}

type webhookEventRepo struct {
	q sqlc.Querier
}

// NewWebhookEventRepo creates a webhook event repository backed by the given querier.
func NewWebhookEventRepo(q sqlc.Querier) WebhookEventRepo {
	return &webhookEventRepo{q: q}
}

func (r *webhookEventRepo) Create(ctx context.Context, integrationID uuid.UUID, providerEventID, eventType string, payload []byte) (sqlc.WebhookEvent, error) {
	event, err := r.q.CreateWebhookEvent(ctx, sqlc.CreateWebhookEventParams{
		IntegrationID:   integrationID,
		ProviderEventID: providerEventID,
		EventType:       eventType,
		Payload:         payload,
	})
	if err != nil {
		return sqlc.WebhookEvent{}, wrapError(err)
	}
	return event, nil
}

func (r *webhookEventRepo) GetByIntegrationAndProvider(ctx context.Context, integrationID uuid.UUID, providerEventID string) (sqlc.WebhookEvent, error) {
	event, err := r.q.GetWebhookEventByIntegrationAndProvider(ctx, sqlc.GetWebhookEventByIntegrationAndProviderParams{
		IntegrationID:   integrationID,
		ProviderEventID: providerEventID,
	})
	if err != nil {
		return sqlc.WebhookEvent{}, wrapNotFound(err)
	}
	return event, nil
}
