package repo

import (
	"context"

	"github.com/I-Frostbyte/rvpay-go/transactions/db/sqlc"
	"github.com/google/uuid"
)

// MerchantRepo provides persistence operations for merchants.
type MerchantRepo interface {
	Create(ctx context.Context, name, slug string, status sqlc.MerchantStatus) (sqlc.Merchant, error)
	GetByID(ctx context.Context, id uuid.UUID) (sqlc.Merchant, error)
	GetBySlug(ctx context.Context, slug string) (sqlc.Merchant, error)
	List(ctx context.Context, limit, offset int32) ([]sqlc.Merchant, error)
	Count(ctx context.Context) (int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.MerchantStatus) (sqlc.Merchant, error)
}

type merchantRepo struct {
	q sqlc.Querier
}

// NewMerchantRepo creates a merchant repository backed by the given querier.
func NewMerchantRepo(q sqlc.Querier) MerchantRepo {
	return &merchantRepo{q: q}
}

func (r *merchantRepo) Create(ctx context.Context, name, slug string, status sqlc.MerchantStatus) (sqlc.Merchant, error) {
	merchant, err := r.q.CreateMerchant(ctx, sqlc.CreateMerchantParams{
		Name:   name,
		Slug:   slug,
		Status: status,
	})
	if err != nil {
		return sqlc.Merchant{}, wrapError(err)
	}
	return merchant, nil
}

func (r *merchantRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Merchant, error) {
	merchant, err := r.q.GetMerchantByID(ctx, id)
	if err != nil {
		return sqlc.Merchant{}, wrapNotFound(err)
	}
	return merchant, nil
}

func (r *merchantRepo) GetBySlug(ctx context.Context, slug string) (sqlc.Merchant, error) {
	merchant, err := r.q.GetMerchantBySlug(ctx, slug)
	if err != nil {
		return sqlc.Merchant{}, wrapNotFound(err)
	}
	return merchant, nil
}

func (r *merchantRepo) List(ctx context.Context, limit, offset int32) ([]sqlc.Merchant, error) {
	merchants, err := r.q.ListMerchants(ctx, sqlc.ListMerchantsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, wrapError(err)
	}
	return merchants, nil
}

func (r *merchantRepo) Count(ctx context.Context) (int64, error) {
	count, err := r.q.CountMerchants(ctx)
	if err != nil {
		return 0, wrapError(err)
	}
	return count, nil
}

func (r *merchantRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.MerchantStatus) (sqlc.Merchant, error) {
	merchant, err := r.q.UpdateMerchantStatus(ctx, sqlc.UpdateMerchantStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return sqlc.Merchant{}, wrapNotFound(err)
	}
	return merchant, nil
}