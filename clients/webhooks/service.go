package webhooks

import (
	"context"
	"time"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/I-Frostbyte/rvpay-go/clients/providers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Service manages webhook lifecycle for provider integrations.
type Service struct {
	integrationsRepo repo.IntegrationRepo
	webhookRepo      repo.WebhookSubscriptionRepo
	platformsRepo    repo.PlatformRepo
	registry         providers.ProviderRegistry
	logger           Logger
}

// Logger defines the logging interface.
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// NewService creates a new webhook service.
func NewService(
	integrationsRepo repo.IntegrationRepo,
	webhookRepo repo.WebhookSubscriptionRepo,
	platformsRepo repo.PlatformRepo,
	registry providers.ProviderRegistry,
	logger Logger,
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

	s.logger.Info("Webhook registered", "integration_id", integrationID, "callback_url", callbackURL)

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

	s.logger.Info("Webhook unregistered", "integration_id", integrationID)

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
		s.logger.Error("Webhook verification failed", "error", err, "provider", providerID)
		return ErrInvalidSignature
	}

	event, err := webhookProvider.ParseEvent(ctx, body)
	if err != nil {
		s.logger.Error("Webhook payload parsing failed", "error", err, "provider", providerID)
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
	s.logger.Info("Webhook subscription exists, processing event", "event_id", event.ProviderEventID, "integration_id", integrationID)

	lastDelivery := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	_, err = s.webhookRepo.UpdateLastDelivery(ctx, webhook.ID, lastDelivery)
	if err != nil {
		return translateError(err)
	}

	dispatcher, ok := provider.(providers.WebhookDispatcher)
	if ok {
		err = dispatcher.Dispatch(ctx, event)
		if err != nil {
			s.logger.Error("Webhook event dispatch failed", "error", err, "event_id", event.ProviderEventID)
			return err
		}
	}

	s.logger.Info("Webhook processed successfully", "provider", providerID, "event_type", event.EventType, "event_id", event.ProviderEventID)

	return nil
}