package repo

import (
	"context"
	"time"

	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/google/uuid"
)

// OAuthStateRepo provides persistence operations for OAuth state records.
type OAuthStateRepo interface {
	Create(ctx context.Context, state string, clientID, platformID uuid.UUID, expiresAt time.Time) (sqlc.OauthState, error)
	GetByState(ctx context.Context, state string) (sqlc.OauthState, error)
	// Consume atomically marks a state as consumed if it is still valid
	// (not consumed and not expired). It returns ErrNotFound when the state
	// is missing, already consumed, or expired.
	Consume(ctx context.Context, state string) (sqlc.OauthState, error)
	DeleteExpired(ctx context.Context) (int64, error)
}

type oauthStateRepo struct {
	q sqlc.Querier
}

// NewOAuthStateRepo creates an OAuth state repository backed by the given querier.
func NewOAuthStateRepo(q sqlc.Querier) OAuthStateRepo {
	return &oauthStateRepo{q: q}
}

func (r *oauthStateRepo) Create(ctx context.Context, state string, clientID, platformID uuid.UUID, expiresAt time.Time) (sqlc.OauthState, error) {
	record, err := r.q.CreateOAuthState(ctx, sqlc.CreateOAuthStateParams{
		State:      state,
		ClientID:   clientID,
		PlatformID: platformID,
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		return sqlc.OauthState{}, wrapError(err)
	}
	return record, nil
}

func (r *oauthStateRepo) GetByState(ctx context.Context, state string) (sqlc.OauthState, error) {
	record, err := r.q.GetOAuthStateByState(ctx, state)
	if err != nil {
		return sqlc.OauthState{}, wrapNotFound(err)
	}
	return record, nil
}

func (r *oauthStateRepo) Consume(ctx context.Context, state string) (sqlc.OauthState, error) {
	record, err := r.q.ConsumeOAuthState(ctx, state)
	if err != nil {
		return sqlc.OauthState{}, wrapNotFound(err)
	}
	return record, nil
}

func (r *oauthStateRepo) DeleteExpired(ctx context.Context) (int64, error) {
	rows, err := r.q.DeleteExpiredOAuthStates(ctx)
	if err != nil {
		return 0, wrapError(err)
	}
	return rows, nil
}
