package repo

import (
	"context"

	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/google/uuid"
)

// PaymentProviderConfigRepo provides persistence operations for GHL Custom
// Payment Provider configurations. Each integration has at most one payment
// provider configuration, keyed by integration_id.
type PaymentProviderConfigRepo interface {
	Create(ctx context.Context, integrationID uuid.UUID, providerName, providerDescription, providerImageURL, locationID, queryURL, paymentsURL string, supportsSubscriptionSchedule bool, providerAPIKey string) (sqlc.PaymentProviderConfig, error)
	GetByIntegrationID(ctx context.Context, integrationID uuid.UUID) (sqlc.PaymentProviderConfig, error)
	GetByLocationID(ctx context.Context, locationID string) (sqlc.PaymentProviderConfig, error)
	GetByAPIKey(ctx context.Context, apiKey string) (sqlc.PaymentProviderConfig, error)
	Update(ctx context.Context, integrationID uuid.UUID, providerName, providerDescription, providerImageURL, locationID, queryURL, paymentsURL string, supportsSubscriptionSchedule bool, providerAPIKey string) (sqlc.PaymentProviderConfig, error)
	Delete(ctx context.Context, integrationID uuid.UUID) error
}

type paymentProviderConfigRepo struct {
	q sqlc.Querier
}

// NewPaymentProviderConfigRepo creates a payment provider config repository
// backed by the given querier.
func NewPaymentProviderConfigRepo(q sqlc.Querier) PaymentProviderConfigRepo {
	return &paymentProviderConfigRepo{q: q}
}

func (r *paymentProviderConfigRepo) Create(ctx context.Context, integrationID uuid.UUID, providerName, providerDescription, providerImageURL, locationID, queryURL, paymentsURL string, supportsSubscriptionSchedule bool, providerAPIKey string) (sqlc.PaymentProviderConfig, error) {
	config, err := r.q.CreatePaymentProviderConfig(ctx, sqlc.CreatePaymentProviderConfigParams{
		IntegrationID:                integrationID,
		ProviderName:                 providerName,
		ProviderDescription:          providerDescription,
		ProviderImageUrl:             providerImageURL,
		LocationID:                   locationID,
		QueryUrl:                     queryURL,
		PaymentsUrl:                  paymentsURL,
		SupportsSubscriptionSchedule: supportsSubscriptionSchedule,
		ProviderApiKey:               providerAPIKey,
	})
	if err != nil {
		return sqlc.PaymentProviderConfig{}, wrapError(err)
	}
	return config, nil
}

func (r *paymentProviderConfigRepo) GetByIntegrationID(ctx context.Context, integrationID uuid.UUID) (sqlc.PaymentProviderConfig, error) {
	config, err := r.q.GetPaymentProviderConfigByIntegrationID(ctx, integrationID)
	if err != nil {
		return sqlc.PaymentProviderConfig{}, wrapNotFound(err)
	}
	return config, nil
}

func (r *paymentProviderConfigRepo) GetByLocationID(ctx context.Context, locationID string) (sqlc.PaymentProviderConfig, error) {
	config, err := r.q.GetPaymentProviderConfigByLocationID(ctx, locationID)
	if err != nil {
		return sqlc.PaymentProviderConfig{}, wrapNotFound(err)
	}
	return config, nil
}

func (r *paymentProviderConfigRepo) GetByAPIKey(ctx context.Context, apiKey string) (sqlc.PaymentProviderConfig, error) {
	config, err := r.q.GetPaymentProviderConfigByAPIKey(ctx, apiKey)
	if err != nil {
		return sqlc.PaymentProviderConfig{}, wrapNotFound(err)
	}
	return config, nil
}

func (r *paymentProviderConfigRepo) Update(ctx context.Context, integrationID uuid.UUID, providerName, providerDescription, providerImageURL, locationID, queryURL, paymentsURL string, supportsSubscriptionSchedule bool, providerAPIKey string) (sqlc.PaymentProviderConfig, error) {
	config, err := r.q.UpdatePaymentProviderConfig(ctx, sqlc.UpdatePaymentProviderConfigParams{
		IntegrationID:                integrationID,
		ProviderName:                 providerName,
		ProviderDescription:          providerDescription,
		ProviderImageUrl:             providerImageURL,
		LocationID:                   locationID,
		QueryUrl:                     queryURL,
		PaymentsUrl:                  paymentsURL,
		SupportsSubscriptionSchedule: supportsSubscriptionSchedule,
		ProviderApiKey:               providerAPIKey,
	})
	if err != nil {
		return sqlc.PaymentProviderConfig{}, wrapNotFound(err)
	}
	return config, nil
}

func (r *paymentProviderConfigRepo) Delete(ctx context.Context, integrationID uuid.UUID) error {
	rows, err := r.q.DeletePaymentProviderConfig(ctx, integrationID)
	if err != nil {
		return wrapError(err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
