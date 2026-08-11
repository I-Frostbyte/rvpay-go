package service

import (
	"context"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	clientsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/clientsgrpc"
	"github.com/rs/zerolog"
)

type PlatformsServiceImpl struct {
	platformsRepo repo.PlatformRepo
	logger        zerolog.Logger
	clientsgrpc.UnimplementedPlatformsServiceServer
}

func NewPlatformsServiceImpl(platformsRepo repo.PlatformRepo, logger zerolog.Logger) *PlatformsServiceImpl {
	return &PlatformsServiceImpl{
		platformsRepo: platformsRepo,
		logger:        logger,
	}
}

func (s *PlatformsServiceImpl) ListPlatforms(ctx context.Context, req *clientsgrpc.ListPlatformsRequest) (*clientsgrpc.ListPlatformsResponse, error) {
	pageSize := int32(20)
	offset := int32(0)
	if req.GetPagination() != nil {
		if req.GetPagination().PageSize > 0 {
			pageSize = req.GetPagination().PageSize
		}
		if req.GetPagination().PageToken != "" {
			offset = pageSize
		}
	}

	platforms, err := s.platformsRepo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, translateRepoError(err)
	}

	protoPlatforms := make([]*clientsgrpc.Platform, 0, len(platforms))
	for _, p := range platforms {
		protoPlatforms = append(protoPlatforms, sqlcPlatformToProto(p))
	}

	return &clientsgrpc.ListPlatformsResponse{
		Platforms: protoPlatforms,
	}, nil
}

func (s *PlatformsServiceImpl) GetPlatform(ctx context.Context, req *clientsgrpc.GetPlatformRequest) (*clientsgrpc.GetPlatformResponse, error) {
	id, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	platform, err := s.platformsRepo.GetByID(ctx, id)
	if err == repo.ErrNotFound {
		return nil, ErrPlatformNotFound
	}
	if err != nil {
		return nil, translateRepoError(err)
	}

	return &clientsgrpc.GetPlatformResponse{
		Platform: sqlcPlatformToProto(platform),
	}, nil
}

func (s *PlatformsServiceImpl) EnablePlatform(ctx context.Context, req *clientsgrpc.EnablePlatformRequest) (*clientsgrpc.EnablePlatformResponse, error) {
	id, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	platform, err := s.platformsRepo.GetByID(ctx, id)
	if err == repo.ErrNotFound {
		return nil, ErrPlatformNotFound
	}
	if err != nil {
		return nil, translateRepoError(err)
	}

	if platform.Enabled {
		return &clientsgrpc.EnablePlatformResponse{
			Platform: sqlcPlatformToProto(platform),
		}, nil
	}

	updated, err := s.platformsRepo.Update(ctx, id, platform.Name, platform.DisplayName, platform.Slug, true, platform.OauthCapable, platform.WebhookCapable)
	if err != nil {
		return nil, translateRepoError(err)
	}

	s.logger.Info().Str("platform_id", updated.ID.String()).Msg("platform enabled")

	return &clientsgrpc.EnablePlatformResponse{
		Platform: sqlcPlatformToProto(updated),
	}, nil
}

func (s *PlatformsServiceImpl) DisablePlatform(ctx context.Context, req *clientsgrpc.DisablePlatformRequest) (*clientsgrpc.DisablePlatformResponse, error) {
	id, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	platform, err := s.platformsRepo.GetByID(ctx, id)
	if err == repo.ErrNotFound {
		return nil, ErrPlatformNotFound
	}
	if err != nil {
		return nil, translateRepoError(err)
	}

	if !platform.Enabled {
		return &clientsgrpc.DisablePlatformResponse{
			Platform: sqlcPlatformToProto(platform),
		}, nil
	}

	updated, err := s.platformsRepo.Update(ctx, id, platform.Name, platform.DisplayName, platform.Slug, false, platform.OauthCapable, platform.WebhookCapable)
	if err != nil {
		return nil, translateRepoError(err)
	}

	s.logger.Info().Str("platform_id", updated.ID.String()).Msg("platform disabled")

	return &clientsgrpc.DisablePlatformResponse{
		Platform: sqlcPlatformToProto(updated),
	}, nil
}
