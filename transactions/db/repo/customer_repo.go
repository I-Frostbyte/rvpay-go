package repo

import (
	"context"

	"github.com/I-Frostbyte/rvpay-go/transactions/db/sqlc"
	"github.com/google/uuid"
)

// CustomerRepo provides persistence operations for customers.
type CustomerRepo interface {
	Create(ctx context.Context, clientID, merchantID uuid.UUID, phoneNumber string, status sqlc.CustomerStatus) (sqlc.Customer, error)
	GetByID(ctx context.Context, id uuid.UUID) (sqlc.Customer, error)
	GetByClientAndMerchantAndPhone(ctx context.Context, clientID, merchantID uuid.UUID, phoneNumber string) (sqlc.Customer, error)
	ListByClient(ctx context.Context, clientID uuid.UUID) ([]sqlc.Customer, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]sqlc.Customer, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.CustomerStatus) (sqlc.Customer, error)
}

type customerRepo struct {
	q sqlc.Querier
}

// NewCustomerRepo creates a customer repository backed by the given querier.
func NewCustomerRepo(q sqlc.Querier) CustomerRepo {
	return &customerRepo{q: q}
}

func (r *customerRepo) Create(ctx context.Context, clientID, merchantID uuid.UUID, phoneNumber string, status sqlc.CustomerStatus) (sqlc.Customer, error) {
	customer, err := r.q.CreateCustomer(ctx, sqlc.CreateCustomerParams{
		ClientID:    clientID,
		MerchantID:  merchantID,
		PhoneNumber: phoneNumber,
		Status:      status,
	})
	if err != nil {
		return sqlc.Customer{}, wrapError(err)
	}
	return customer, nil
}

func (r *customerRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Customer, error) {
	customer, err := r.q.GetCustomerByID(ctx, id)
	if err != nil {
		return sqlc.Customer{}, wrapNotFound(err)
	}
	return customer, nil
}

func (r *customerRepo) GetByClientAndMerchantAndPhone(ctx context.Context, clientID, merchantID uuid.UUID, phoneNumber string) (sqlc.Customer, error) {
	customer, err := r.q.GetCustomerByClientAndMerchantAndPhone(ctx, sqlc.GetCustomerByClientAndMerchantAndPhoneParams{
		ClientID:    clientID,
		MerchantID:  merchantID,
		PhoneNumber: phoneNumber,
	})
	if err != nil {
		return sqlc.Customer{}, wrapNotFound(err)
	}
	return customer, nil
}

func (r *customerRepo) ListByClient(ctx context.Context, clientID uuid.UUID) ([]sqlc.Customer, error) {
	customers, err := r.q.ListCustomersByClient(ctx, clientID)
	if err != nil {
		return nil, wrapError(err)
	}
	return customers, nil
}

func (r *customerRepo) ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]sqlc.Customer, error) {
	customers, err := r.q.ListCustomersByMerchant(ctx, merchantID)
	if err != nil {
		return nil, wrapError(err)
	}
	return customers, nil
}

func (r *customerRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.CustomerStatus) (sqlc.Customer, error) {
	customer, err := r.q.UpdateCustomerStatus(ctx, sqlc.UpdateCustomerStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return sqlc.Customer{}, wrapNotFound(err)
	}
	return customer, nil
}
