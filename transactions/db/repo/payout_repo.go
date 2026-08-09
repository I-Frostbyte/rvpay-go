package repo

import (
	"context"

	"github.com/I-Frostbyte/rvpay-go/transactions/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// PayoutRepo provides persistence operations for payouts.
type PayoutRepo interface {
	Create(ctx context.Context, clientID, merchantID uuid.UUID, amount pgtype.Numeric, currency string, provider sqlc.PaymentProvider, destinationReference string, status sqlc.PayoutStatus, idempotencyKey uuid.UUID) (sqlc.Payout, error)
	GetByID(ctx context.Context, id uuid.UUID) (sqlc.Payout, error)
	GetByExternalReference(ctx context.Context, externalReference string) (sqlc.Payout, error)
	GetByIdempotencyKey(ctx context.Context, idempotencyKey uuid.UUID) (sqlc.Payout, error)
	ListByClient(ctx context.Context, clientID uuid.UUID) ([]sqlc.Payout, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]sqlc.Payout, error)
	ListByStatus(ctx context.Context, status sqlc.PayoutStatus) ([]sqlc.Payout, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.PayoutStatus) (sqlc.Payout, error)
	MarkCompleted(ctx context.Context, id uuid.UUID, status sqlc.PayoutStatus) (sqlc.Payout, error)
	MarkFailed(ctx context.Context, id uuid.UUID, status sqlc.PayoutStatus, failureReason string) (sqlc.Payout, error)
}

type payoutRepo struct {
	q sqlc.Querier
}

// NewPayoutRepo creates a payout repository backed by the given querier.
func NewPayoutRepo(q sqlc.Querier) PayoutRepo {
	return &payoutRepo{q: q}
}

func (r *payoutRepo) Create(ctx context.Context, clientID, merchantID uuid.UUID, amount pgtype.Numeric, currency string, provider sqlc.PaymentProvider, destinationReference string, status sqlc.PayoutStatus, idempotencyKey uuid.UUID) (sqlc.Payout, error) {
	payout, err := r.q.CreatePayout(ctx, sqlc.CreatePayoutParams{
		ClientID:             clientID,
		MerchantID:           merchantID,
		Amount:               amount,
		Currency:             currency,
		Provider:             provider,
		DestinationReference: destinationReference,
		Status:               status,
		IdempotencyKey:       idempotencyKey,
	})
	if err != nil {
		return sqlc.Payout{}, wrapError(err)
	}
	return payout, nil
}

func (r *payoutRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Payout, error) {
	payout, err := r.q.GetPayoutByID(ctx, id)
	if err != nil {
		return sqlc.Payout{}, wrapNotFound(err)
	}
	return payout, nil
}

func (r *payoutRepo) GetByExternalReference(ctx context.Context, externalReference string) (sqlc.Payout, error) {
	payout, err := r.q.GetPayoutByExternalReference(ctx, externalReference)
	if err != nil {
		return sqlc.Payout{}, wrapNotFound(err)
	}
	return payout, nil
}

func (r *payoutRepo) GetByIdempotencyKey(ctx context.Context, idempotencyKey uuid.UUID) (sqlc.Payout, error) {
	payout, err := r.q.GetPayoutByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return sqlc.Payout{}, wrapNotFound(err)
	}
	return payout, nil
}

func (r *payoutRepo) ListByClient(ctx context.Context, clientID uuid.UUID) ([]sqlc.Payout, error) {
	payouts, err := r.q.ListPayoutsByClient(ctx, clientID)
	if err != nil {
		return nil, wrapError(err)
	}
	return payouts, nil
}

func (r *payoutRepo) ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]sqlc.Payout, error) {
	payouts, err := r.q.ListPayoutsByMerchant(ctx, merchantID)
	if err != nil {
		return nil, wrapError(err)
	}
	return payouts, nil
}

func (r *payoutRepo) ListByStatus(ctx context.Context, status sqlc.PayoutStatus) ([]sqlc.Payout, error) {
	payouts, err := r.q.ListPayoutsByStatus(ctx, status)
	if err != nil {
		return nil, wrapError(err)
	}
	return payouts, nil
}

func (r *payoutRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.PayoutStatus) (sqlc.Payout, error) {
	payout, err := r.q.UpdatePayoutStatus(ctx, sqlc.UpdatePayoutStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return sqlc.Payout{}, wrapNotFound(err)
	}
	return payout, nil
}

func (r *payoutRepo) MarkCompleted(ctx context.Context, id uuid.UUID, status sqlc.PayoutStatus) (sqlc.Payout, error) {
	payout, err := r.q.UpdatePayoutStatusAndCompletedAt(ctx, sqlc.UpdatePayoutStatusAndCompletedAtParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return sqlc.Payout{}, wrapNotFound(err)
	}
	return payout, nil
}

func (r *payoutRepo) MarkFailed(ctx context.Context, id uuid.UUID, status sqlc.PayoutStatus, failureReason string) (sqlc.Payout, error) {
	payout, err := r.q.UpdatePayoutStatusAndFailedAt(ctx, sqlc.UpdatePayoutStatusAndFailedAtParams{
		ID:            id,
		Status:        status,
		FailureReason: failureReason,
	})
	if err != nil {
		return sqlc.Payout{}, wrapNotFound(err)
	}
	return payout, nil
}