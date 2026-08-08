package repo

import (
	"context"

	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// IntegrationRepo provides persistence operations for integrations.
type IntegrationRepo interface {
	Create(ctx context.Context, clientID, platformID uuid.UUID, externalAccountID string, status sqlc.IntegrationStatus) (sqlc.Integration, error)
	GetByID(ctx context.Context, id uuid.UUID) (sqlc.Integration, error)
	GetByClientAndPlatform(ctx context.Context, clientID, platformID uuid.UUID) (sqlc.Integration, error)
	GetByExternalAccountID(ctx context.Context, externalAccountID string) (sqlc.Integration, error)
	ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error)
	ListByPlatform(ctx context.Context, platformID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error)
	ListActiveByClient(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error)
	CountByClient(ctx context.Context, clientID uuid.UUID) (int64, error)
	ExistsByClientAndPlatform(ctx context.Context, clientID, platformID uuid.UUID) (bool, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.IntegrationStatus) (sqlc.Integration, error)
	UpdateLastSyncAt(ctx context.Context, id uuid.UUID, lastSyncAt pgtype.Timestamptz) (sqlc.Integration, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type integrationRepo struct {
	q sqlc.Querier
}

// NewIntegrationRepo creates an integration repository backed by the given querier.
func NewIntegrationRepo(q sqlc.Querier) IntegrationRepo {
	return &integrationRepo{q: q}
}

func (r *integrationRepo) Create(ctx context.Context, clientID, platformID uuid.UUID, externalAccountID string, status sqlc.IntegrationStatus) (sqlc.Integration, error) {
	integration, err := r.q.CreateIntegration(ctx, sqlc.CreateIntegrationParams{
		ClientID:          clientID,
		PlatformID:        platformID,
		ExternalAccountID: externalAccountID,
		Status:            status,
	})
	if err != nil {
		return sqlc.Integration{}, wrapError(err)
	}
	return integration, nil
}

func (r *integrationRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Integration, error) {
	integration, err := r.q.GetIntegrationByID(ctx, id)
	if err != nil {
		return sqlc.Integration{}, wrapNotFound(err)
	}
	return integration, nil
}

func (r *integrationRepo) GetByClientAndPlatform(ctx context.Context, clientID, platformID uuid.UUID) (sqlc.Integration, error) {
	integration, err := r.q.GetIntegrationByClientAndPlatform(ctx, sqlc.GetIntegrationByClientAndPlatformParams{
		ClientID:   clientID,
		PlatformID: platformID,
	})
	if err != nil {
		return sqlc.Integration{}, wrapNotFound(err)
	}
	return integration, nil
}

func (r *integrationRepo) GetByExternalAccountID(ctx context.Context, externalAccountID string) (sqlc.Integration, error) {
	integration, err := r.q.GetIntegrationByExternalAccountID(ctx, externalAccountID)
	if err != nil {
		return sqlc.Integration{}, wrapNotFound(err)
	}
	return integration, nil
}

func (r *integrationRepo) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	integrations, err := r.q.ListIntegrationsByClient(ctx, sqlc.ListIntegrationsByClientParams{
		ClientID: clientID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, wrapError(err)
	}
	return integrations, nil
}

func (r *integrationRepo) ListByPlatform(ctx context.Context, platformID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	integrations, err := r.q.ListIntegrationsByPlatform(ctx, sqlc.ListIntegrationsByPlatformParams{
		PlatformID: platformID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, wrapError(err)
	}
	return integrations, nil
}

func (r *integrationRepo) ListActiveByClient(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]sqlc.Integration, error) {
	integrations, err := r.q.ListActiveIntegrationsByClient(ctx, sqlc.ListActiveIntegrationsByClientParams{
		ClientID: clientID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, wrapError(err)
	}
	return integrations, nil
}

func (r *integrationRepo) CountByClient(ctx context.Context, clientID uuid.UUID) (int64, error) {
	count, err := r.q.CountIntegrationsByClient(ctx, clientID)
	if err != nil {
		return 0, wrapError(err)
	}
	return count, nil
}

func (r *integrationRepo) ExistsByClientAndPlatform(ctx context.Context, clientID, platformID uuid.UUID) (bool, error) {
	exists, err := r.q.IntegrationExistsByClientAndPlatform(ctx, sqlc.IntegrationExistsByClientAndPlatformParams{
		ClientID:   clientID,
		PlatformID: platformID,
	})
	if err != nil {
		return false, wrapError(err)
	}
	return exists, nil
}

func (r *integrationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.IntegrationStatus) (sqlc.Integration, error) {
	integration, err := r.q.UpdateIntegrationStatus(ctx, sqlc.UpdateIntegrationStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return sqlc.Integration{}, wrapNotFound(err)
	}
	return integration, nil
}

func (r *integrationRepo) UpdateLastSyncAt(ctx context.Context, id uuid.UUID, lastSyncAt pgtype.Timestamptz) (sqlc.Integration, error) {
	integration, err := r.q.UpdateIntegrationLastSyncAt(ctx, sqlc.UpdateIntegrationLastSyncAtParams{
		ID:         id,
		LastSyncAt: lastSyncAt,
	})
	if err != nil {
		return sqlc.Integration{}, wrapNotFound(err)
	}
	return integration, nil
}

func (r *integrationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.q.DeleteIntegration(ctx, id)
	if err != nil {
		return wrapError(err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}