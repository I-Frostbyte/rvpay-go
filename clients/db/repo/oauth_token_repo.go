package repo

import (
	"context"
	"time"

	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/google/uuid"
)

// OAuthTokenRepo provides persistence operations for OAuth tokens.
type OAuthTokenRepo interface {
	Create(ctx context.Context, integrationID uuid.UUID, accessToken, refreshToken string, expiresAt time.Time, scope, tokenType string) (sqlc.OauthToken, error)
	GetByID(ctx context.Context, id uuid.UUID) (sqlc.OauthToken, error)
	GetByIntegrationID(ctx context.Context, integrationID uuid.UUID) (sqlc.OauthToken, error)
	ExistsByIntegrationID(ctx context.Context, integrationID uuid.UUID) (bool, error)
	Update(ctx context.Context, id uuid.UUID, accessToken, refreshToken string, expiresAt time.Time, scope, tokenType string) (sqlc.OauthToken, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByIntegrationID(ctx context.Context, integrationID uuid.UUID) error
}

type oauthTokenRepo struct {
	q sqlc.Querier
}

// NewOAuthTokenRepo creates an OAuth token repository backed by the given querier.
func NewOAuthTokenRepo(q sqlc.Querier) OAuthTokenRepo {
	return &oauthTokenRepo{q: q}
}

func (r *oauthTokenRepo) Create(ctx context.Context, integrationID uuid.UUID, accessToken, refreshToken string, expiresAt time.Time, scope, tokenType string) (sqlc.OauthToken, error) {
	token, err := r.q.CreateOAuthToken(ctx, sqlc.CreateOAuthTokenParams{
		IntegrationID: integrationID,
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		ExpiresAt:     expiresAt,
		Scope:         scope,
		TokenType:     tokenType,
	})
	if err != nil {
		return sqlc.OauthToken{}, wrapError(err)
	}
	return token, nil
}

func (r *oauthTokenRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.OauthToken, error) {
	token, err := r.q.GetOAuthTokenByID(ctx, id)
	if err != nil {
		return sqlc.OauthToken{}, wrapNotFound(err)
	}
	return token, nil
}

func (r *oauthTokenRepo) GetByIntegrationID(ctx context.Context, integrationID uuid.UUID) (sqlc.OauthToken, error) {
	token, err := r.q.GetOAuthTokenByIntegrationID(ctx, integrationID)
	if err != nil {
		return sqlc.OauthToken{}, wrapNotFound(err)
	}
	return token, nil
}

func (r *oauthTokenRepo) ExistsByIntegrationID(ctx context.Context, integrationID uuid.UUID) (bool, error) {
	exists, err := r.q.OAuthTokenExistsByIntegrationID(ctx, integrationID)
	if err != nil {
		return false, wrapError(err)
	}
	return exists, nil
}

func (r *oauthTokenRepo) Update(ctx context.Context, id uuid.UUID, accessToken, refreshToken string, expiresAt time.Time, scope, tokenType string) (sqlc.OauthToken, error) {
	token, err := r.q.UpdateOAuthToken(ctx, sqlc.UpdateOAuthTokenParams{
		ID:           id,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		Scope:        scope,
		TokenType:    tokenType,
	})
	if err != nil {
		return sqlc.OauthToken{}, wrapNotFound(err)
	}
	return token, nil
}

func (r *oauthTokenRepo) Delete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.q.DeleteOAuthToken(ctx, id)
	if err != nil {
		return wrapError(err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *oauthTokenRepo) DeleteByIntegrationID(ctx context.Context, integrationID uuid.UUID) error {
	_, err := r.q.DeleteOAuthTokenByIntegrationID(ctx, integrationID)
	if err != nil {
		return wrapError(err)
	}
	return nil
}
