package webhooks

import (
	"context"
	"time"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/I-Frostbyte/rvpay-go/clients/providers"
	"github.com/rs/zerolog"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Service manages webhook lifecycle for provider integrations.
type Service struct {
	integrationsRepo repo.IntegrationRepo
	webhookRepo      repo.WebhookSubscriptionRepo
	platformsRepo    repo.PlatformRepo
	registry         providers.ProviderRegistry
	logger           zerolog.Logger
}

// NewService creates a new webhook service.
func NewService(
	integrationsRepo repo.IntegrationRepo,
	webhookRepo repo.WebhookSubscriptionRepo,
	platformsRepo repo.PlatformRepo,
	registry providers.ProviderRegistry,
	logger zerolog.Logger,
) *Service {
	return &Service{
		integrationsRepo: integrationsRepo,
		webhookRepo:      webhookRepo,
		platformsRepo:    platformsRepo,
		registry:         registry,
		logger:           logger,
	}
}

// RegisterWebhook registers a webhook subscription for an integration.
func (s *Service) RegisterWebhook(ctx context.Context, integrationID uuid.UUID, callbackURL string) error {
	integration, err := s.integrationsRepo.GetByID(ctx, integrationID)
	if err == repo.ErrNotFound {
		return ErrIntegrationNotFound
	}
	if err != nil {
		return translateError(err)
	}

	if integration.Status != sqlc.IntegrationStatusACTIVE {
		return ErrIntegrationNotActive
	}

	platform, err := s.platformsRepo.GetByID(ctx, integration.PlatformID)
	if err == repo.ErrNotFound {
		return ErrPlatformNotFound
	}
	if err != nil {
		return translateError(err)
	}

	provider, ok := s.registry.Get(platform.Slug)
	if !ok {
		return ErrProviderNotSupported
	}

	webhookProvider, ok := provider.(providers.WebhookProvider)
	if !ok {
		return ErrProviderNotSupported
	}

	err = webhookProvider.RegisterWebhook(ctx, integrationID.String(), callbackURL)
	if err != nil {
		return err
	}

	_, err = s.webhookRepo.Create(ctx, integrationID, callbackURL, "highlevel", sqlc.WebhookSubscriptionStatusACTIVE)
	if err != nil {
		return translateError(err)
	}

	s.logger.Info().Str("integration_id", integrationID.String()).Str("callback_url", callbackURL).Msg("Webhook registered")

	return nil
}

// UnregisterWebhook removes a webhook subscription for an integration.
func (s *Service) UnregisterWebhook(ctx context.Context, integrationID uuid.UUID) error {
	webhook, err := s.webhookRepo.GetByIntegrationIDAndEndpoint(ctx, integrationID, "")
	if err == repo.ErrNotFound {
		return ErrWebhookNotFound
	}
	if err != nil {
		return translateError(err)
	}

	integration, err := s.integrationsRepo.GetByID(ctx, integrationID)
	if err == repo.ErrNotFound {
		return ErrIntegrationNotFound
	}
	if err != nil {
		return translateError(err)
	}

	platform, err := s.platformsRepo.GetByID(ctx, integration.PlatformID)
	if err == repo.ErrNotFound {
		return ErrPlatformNotFound
	}
	if err != nil {
		return translateError(err)
	}

	provider, ok := s.registry.Get(platform.Slug)
	if !ok {
		return ErrProviderNotSupported
	}

	webhookProvider, ok := provider.(providers.WebhookProvider)
	if !ok {
		return ErrProviderNotSupported
	}

	err = webhookProvider.UnregisterWebhook(ctx, integrationID.String())
	if err != nil {
		return err
	}

	err = s.webhookRepo.Delete(ctx, webhook.ID)
	if err != nil {
		return translateError(err)
	}

	s.logger.Info().Str("integration_id", integrationID.String()).Msg("Webhook unregistered")

	return nil
}

// ProcessWebhook processes an incoming webhook request.
func (s *Service) ProcessWebhook(ctx context.Context, providerID string, headers map[string]string, body []byte) error {
	provider, ok := s.registry.Get(providerID)
	if !ok {
		return ErrProviderNotSupported
	}

	webhookProvider, ok := provider.(providers.WebhookProvider)
	if !ok {
		return ErrProviderNotSupported
	}

	err := webhookProvider.VerifyRequest(ctx, headers, body)
	if err != nil {
		s.logger.Error().Err(err).Str("provider", providerID).Msg("Webhook verification failed")
		return ErrInvalidSignature
	}

	event, err := webhookProvider.ParseEvent(ctx, body)
	if err != nil {
		s.logger.Error().Err(err).Str("provider", providerID).Msg("Webhook payload parsing failed")
		return ErrInvalidPayload
	}

	integrationID, err := uuid.Parse(event.IntegrationID)
	if err != nil {
		return ErrInvalidPayload
	}

	webhook, err := s.webhookRepo.GetByIntegrationIDAndEndpoint(ctx, integrationID, "")
	if err == repo.ErrNotFound {
		return ErrWebhookNotFound
	}
	if err != nil {
		return translateError(err)
	}

	// Duplicate detection: if webhook exists, log and continue
	// In production, this would check a webhook_events table for provider_event_id
	s.logger.Info().Str("event_id", event.ProviderEventID).Str("integration_id", integrationID.String()).Msg("Webhook subscription exists, processing event")

	lastDelivery := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	_, err = s.webhookRepo.UpdateLastDelivery(ctx, webhook.ID, lastDelivery)
	if err != nil {
		return translateError(err)
	}

	dispatcher, ok := provider.(providers.WebhookDispatcher)
	if ok {
		err = dispatcher.Dispatch(ctx, event)
		if err != nil {
			s.logger.Error().Err(err).Str("event_id", event.ProviderEventID).Msg("Webhook event dispatch failed")
			return err
		}
	}

	s.logger.Info().Str("provider", providerID).Str("event_type", event.EventType).Str("event_id", event.ProviderEventID).Msg("Webhook processed successfully")

	return nil
}