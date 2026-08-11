package repo

import (
	"context"

	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/google/uuid"
)

// PlatformRepo provides persistence operations for platforms.
type PlatformRepo interface {
	Create(ctx context.Context, name, displayName, slug string, enabled, oauthCapable, webhookCapable bool) (sqlc.Platform, error)
	GetByID(ctx context.Context, id uuid.UUID) (sqlc.Platform, error)
	GetByName(ctx context.Context, name string) (sqlc.Platform, error)
	GetBySlug(ctx context.Context, slug string) (sqlc.Platform, error)
	List(ctx context.Context, limit, offset int32) ([]sqlc.Platform, error)
	ListEnabled(ctx context.Context, limit, offset int32) ([]sqlc.Platform, error)
	Count(ctx context.Context) (int64, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Update(ctx context.Context, id uuid.UUID, name, displayName, slug string, enabled, oauthCapable, webhookCapable bool) (sqlc.Platform, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type platformRepo struct {
	q sqlc.Querier
}

// NewPlatformRepo creates a platform repository backed by the given querier.
func NewPlatformRepo(q sqlc.Querier) PlatformRepo {
	return &platformRepo{q: q}
}

func (r *platformRepo) Create(ctx context.Context, name, displayName, slug string, enabled, oauthCapable, webhookCapable bool) (sqlc.Platform, error) {
	platform, err := r.q.CreatePlatform(ctx, sqlc.CreatePlatformParams{
		Name:           name,
		DisplayName:    displayName,
		Slug:           slug,
		Enabled:        enabled,
		OauthCapable:   oauthCapable,
		WebhookCapable: webhookCapable,
	})
	if err != nil {
		return sqlc.Platform{}, wrapError(err)
	}
	return platform, nil
}

func (r *platformRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Platform, error) {
	platform, err := r.q.GetPlatformByID(ctx, id)
	if err != nil {
		return sqlc.Platform{}, wrapNotFound(err)
	}
	return platform, nil
}

func (r *platformRepo) GetByName(ctx context.Context, name string) (sqlc.Platform, error) {
	platform, err := r.q.GetPlatformByName(ctx, name)
	if err != nil {
		return sqlc.Platform{}, wrapNotFound(err)
	}
	return platform, nil
}

func (r *platformRepo) GetBySlug(ctx context.Context, slug string) (sqlc.Platform, error) {
	platform, err := r.q.GetPlatformBySlug(ctx, slug)
	if err != nil {
		return sqlc.Platform{}, wrapNotFound(err)
	}
	return platform, nil
}

func (r *platformRepo) List(ctx context.Context, limit, offset int32) ([]sqlc.Platform, error) {
	platforms, err := r.q.ListPlatforms(ctx, sqlc.ListPlatformsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, wrapError(err)
	}
	return platforms, nil
}

func (r *platformRepo) ListEnabled(ctx context.Context, limit, offset int32) ([]sqlc.Platform, error) {
	platforms, err := r.q.ListEnabledPlatforms(ctx, sqlc.ListEnabledPlatformsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, wrapError(err)
	}
	return platforms, nil
}

func (r *platformRepo) Count(ctx context.Context) (int64, error) {
	count, err := r.q.CountPlatforms(ctx)
	if err != nil {
		return 0, wrapError(err)
	}
	return count, nil
}

func (r *platformRepo) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	exists, err := r.q.PlatformExistsBySlug(ctx, slug)
	if err != nil {
		return false, wrapError(err)
	}
	return exists, nil
}

func (r *platformRepo) Update(ctx context.Context, id uuid.UUID, name, displayName, slug string, enabled, oauthCapable, webhookCapable bool) (sqlc.Platform, error) {
	platform, err := r.q.UpdatePlatform(ctx, sqlc.UpdatePlatformParams{
		ID:             id,
		Name:           name,
		DisplayName:    displayName,
		Slug:           slug,
		Enabled:        enabled,
		OauthCapable:   oauthCapable,
		WebhookCapable: webhookCapable,
	})
	if err != nil {
		return sqlc.Platform{}, wrapNotFound(err)
	}
	return platform, nil
}

func (r *platformRepo) Delete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.q.DeletePlatform(ctx, id)
	if err != nil {
		return wrapError(err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
