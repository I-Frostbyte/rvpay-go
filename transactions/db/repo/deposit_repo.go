package repo

import (
	"context"

	"github.com/I-Frostbyte/rvpay-go/transactions/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// DepositRepo provides persistence operations for deposits.
type DepositRepo interface {
	Create(ctx context.Context, clientID, customerID, merchantID uuid.UUID, amount pgtype.Numeric, currency string, paymentType sqlc.PaymentType, payerPhoneNumber string, provider sqlc.PaymentProvider, status sqlc.DepositStatus, idempotencyKey uuid.UUID) (sqlc.Deposit, error)
	GetByID(ctx context.Context, id uuid.UUID) (sqlc.Deposit, error)
	GetByExternalReference(ctx context.Context, externalReference string) (sqlc.Deposit, error)
	GetByIdempotencyKey(ctx context.Context, idempotencyKey uuid.UUID) (sqlc.Deposit, error)
	ListByClient(ctx context.Context, clientID uuid.UUID) ([]sqlc.Deposit, error)
	ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]sqlc.Deposit, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]sqlc.Deposit, error)
	ListByStatus(ctx context.Context, status sqlc.DepositStatus) ([]sqlc.Deposit, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.DepositStatus) (sqlc.Deposit, error)
	MarkCompleted(ctx context.Context, id uuid.UUID, status sqlc.DepositStatus) (sqlc.Deposit, error)
	MarkFailed(ctx context.Context, id uuid.UUID, status sqlc.DepositStatus, failureReason string) (sqlc.Deposit, error)
}

type depositRepo struct {
	q sqlc.Querier
}

// NewDepositRepo creates a deposit repository backed by the given querier.
func NewDepositRepo(q sqlc.Querier) DepositRepo {
	return &depositRepo{q: q}
}

func (r *depositRepo) Create(ctx context.Context, clientID, customerID, merchantID uuid.UUID, amount pgtype.Numeric, currency string, paymentType sqlc.PaymentType, payerPhoneNumber string, provider sqlc.PaymentProvider, status sqlc.DepositStatus, idempotencyKey uuid.UUID) (sqlc.Deposit, error) {
	deposit, err := r.q.CreateDeposit(ctx, sqlc.CreateDepositParams{
		ClientID:         clientID,
		CustomerID:       customerID,
		MerchantID:       merchantID,
		Amount:           amount,
		Currency:         currency,
		PaymentType:      paymentType,
		PayerPhoneNumber: payerPhoneNumber,
		Provider:         provider,
		Status:           status,
		IdempotencyKey:   idempotencyKey,
	})
	if err != nil {
		return sqlc.Deposit{}, wrapError(err)
	}
	return deposit, nil
}

func (r *depositRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Deposit, error) {
	deposit, err := r.q.GetDepositByID(ctx, id)
	if err != nil {
		return sqlc.Deposit{}, wrapNotFound(err)
	}
	return deposit, nil
}

func (r *depositRepo) GetByExternalReference(ctx context.Context, externalReference string) (sqlc.Deposit, error) {
	deposit, err := r.q.GetDepositByExternalReference(ctx, externalReference)
	if err != nil {
		return sqlc.Deposit{}, wrapNotFound(err)
	}
	return deposit, nil
}

func (r *depositRepo) GetByIdempotencyKey(ctx context.Context, idempotencyKey uuid.UUID) (sqlc.Deposit, error) {
	deposit, err := r.q.GetDepositByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return sqlc.Deposit{}, wrapNotFound(err)
	}
	return deposit, nil
}

func (r *depositRepo) ListByClient(ctx context.Context, clientID uuid.UUID) ([]sqlc.Deposit, error) {
	deposits, err := r.q.ListDepositsByClient(ctx, clientID)
	if err != nil {
		return nil, wrapError(err)
	}
	return deposits, nil
}

func (r *depositRepo) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]sqlc.Deposit, error) {
	deposits, err := r.q.ListDepositsByCustomer(ctx, customerID)
	if err != nil {
		return nil, wrapError(err)
	}
	return deposits, nil
}

func (r *depositRepo) ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]sqlc.Deposit, error) {
	deposits, err := r.q.ListDepositsByMerchant(ctx, merchantID)
	if err != nil {
		return nil, wrapError(err)
	}
	return deposits, nil
}

func (r *depositRepo) ListByStatus(ctx context.Context, status sqlc.DepositStatus) ([]sqlc.Deposit, error) {
	deposits, err := r.q.ListDepositsByStatus(ctx, status)
	if err != nil {
		return nil, wrapError(err)
	}
	return deposits, nil
}

func (r *depositRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.DepositStatus) (sqlc.Deposit, error) {
	deposit, err := r.q.UpdateDepositStatus(ctx, sqlc.UpdateDepositStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return sqlc.Deposit{}, wrapNotFound(err)
	}
	return deposit, nil
}

func (r *depositRepo) MarkCompleted(ctx context.Context, id uuid.UUID, status sqlc.DepositStatus) (sqlc.Deposit, error) {
	deposit, err := r.q.UpdateDepositStatusAndCompletedAt(ctx, sqlc.UpdateDepositStatusAndCompletedAtParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return sqlc.Deposit{}, wrapNotFound(err)
	}
	return deposit, nil
}

func (r *depositRepo) MarkFailed(ctx context.Context, id uuid.UUID, status sqlc.DepositStatus, failureReason string) (sqlc.Deposit, error) {
	deposit, err := r.q.UpdateDepositStatusAndFailedAt(ctx, sqlc.UpdateDepositStatusAndFailedAtParams{
		ID:            id,
		Status:        status,
		FailureReason: failureReason,
	})
	if err != nil {
		return sqlc.Deposit{}, wrapNotFound(err)
	}
	return deposit, nil
}