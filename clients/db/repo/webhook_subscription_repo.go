package repo

import (
	"context"

	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// WebhookSubscriptionRepo provides persistence operations for webhook subscriptions.
type WebhookSubscriptionRepo interface {
	Create(ctx context.Context, integrationID uuid.UUID, endpoint, secret string, status sqlc.WebhookSubscriptionStatus) (sqlc.WebhookSubscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (sqlc.WebhookSubscription, error)
	GetByIntegrationIDAndEndpoint(ctx context.Context, integrationID uuid.UUID, endpoint string) (sqlc.WebhookSubscription, error)
	ListByIntegrationID(ctx context.Context, integrationID uuid.UUID, limit, offset int32) ([]sqlc.WebhookSubscription, error)
	ListActiveByIntegrationID(ctx context.Context, integrationID uuid.UUID) ([]sqlc.WebhookSubscription, error)
	CountByIntegrationID(ctx context.Context, integrationID uuid.UUID) (int64, error)
	Exists(ctx context.Context, integrationID uuid.UUID, endpoint string) (bool, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.WebhookSubscriptionStatus) (sqlc.WebhookSubscription, error)
	UpdateLastDelivery(ctx context.Context, id uuid.UUID, lastDelivery pgtype.Timestamptz) (sqlc.WebhookSubscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type webhookSubscriptionRepo struct {
	q sqlc.Querier
}

// NewWebhookSubscriptionRepo creates a webhook subscription repository backed by the given querier.
func NewWebhookSubscriptionRepo(q sqlc.Querier) WebhookSubscriptionRepo {
	return &webhookSubscriptionRepo{q: q}
}

func (r *webhookSubscriptionRepo) Create(ctx context.Context, integrationID uuid.UUID, endpoint, secret string, status sqlc.WebhookSubscriptionStatus) (sqlc.WebhookSubscription, error) {
	sub, err := r.q.CreateWebhookSubscription(ctx, sqlc.CreateWebhookSubscriptionParams{
		IntegrationID: integrationID,
		Endpoint:      endpoint,
		Secret:        secret,
		Status:        status,
	})
	if err != nil {
		return sqlc.WebhookSubscription{}, wrapError(err)
	}
	return sub, nil
}

func (r *webhookSubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.WebhookSubscription, error) {
	sub, err := r.q.GetWebhookSubscriptionByID(ctx, id)
	if err != nil {
		return sqlc.WebhookSubscription{}, wrapNotFound(err)
	}
	return sub, nil
}

func (r *webhookSubscriptionRepo) GetByIntegrationIDAndEndpoint(ctx context.Context, integrationID uuid.UUID, endpoint string) (sqlc.WebhookSubscription, error) {
	sub, err := r.q.GetWebhookSubscriptionByIntegrationIDAndEndpoint(ctx, sqlc.GetWebhookSubscriptionByIntegrationIDAndEndpointParams{
		IntegrationID: integrationID,
		Endpoint:      endpoint,
	})
	if err != nil {
		return sqlc.WebhookSubscription{}, wrapNotFound(err)
	}
	return sub, nil
}

func (r *webhookSubscriptionRepo) ListByIntegrationID(ctx context.Context, integrationID uuid.UUID, limit, offset int32) ([]sqlc.WebhookSubscription, error) {
	subs, err := r.q.ListWebhookSubscriptionsByIntegrationID(ctx, sqlc.ListWebhookSubscriptionsByIntegrationIDParams{
		IntegrationID: integrationID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, wrapError(err)
	}
	return subs, nil
}

func (r *webhookSubscriptionRepo) ListActiveByIntegrationID(ctx context.Context, integrationID uuid.UUID) ([]sqlc.WebhookSubscription, error) {
	subs, err := r.q.ListActiveWebhookSubscriptionsByIntegrationID(ctx, integrationID)
	if err != nil {
		return nil, wrapError(err)
	}
	return subs, nil
}

func (r *webhookSubscriptionRepo) CountByIntegrationID(ctx context.Context, integrationID uuid.UUID) (int64, error) {
	count, err := r.q.CountWebhookSubscriptionsByIntegrationID(ctx, integrationID)
	if err != nil {
		return 0, wrapError(err)
	}
	return count, nil
}

func (r *webhookSubscriptionRepo) Exists(ctx context.Context, integrationID uuid.UUID, endpoint string) (bool, error) {
	exists, err := r.q.WebhookSubscriptionExists(ctx, sqlc.WebhookSubscriptionExistsParams{
		IntegrationID: integrationID,
		Endpoint:      endpoint,
	})
	if err != nil {
		return false, wrapError(err)
	}
	return exists, nil
}

func (r *webhookSubscriptionRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.WebhookSubscriptionStatus) (sqlc.WebhookSubscription, error) {
	sub, err := r.q.UpdateWebhookSubscriptionStatus(ctx, sqlc.UpdateWebhookSubscriptionStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return sqlc.WebhookSubscription{}, wrapNotFound(err)
	}
	return sub, nil
}

func (r *webhookSubscriptionRepo) UpdateLastDelivery(ctx context.Context, id uuid.UUID, lastDelivery pgtype.Timestamptz) (sqlc.WebhookSubscription, error) {
	sub, err := r.q.UpdateWebhookSubscriptionLastDelivery(ctx, sqlc.UpdateWebhookSubscriptionLastDeliveryParams{
		ID:           id,
		LastDelivery: lastDelivery,
	})
	if err != nil {
		return sqlc.WebhookSubscription{}, wrapNotFound(err)
	}
	return sub, nil
}

func (r *webhookSubscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.q.DeleteWebhookSubscription(ctx, id)
	if err != nil {
		return wrapError(err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
