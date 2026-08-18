package webhooks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/I-Frostbyte/rvpay-go/clients/providers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

// Service manages webhook lifecycle for provider integrations.
type Service struct {
	integrationsRepo          repo.IntegrationRepo
	webhookRepo               repo.WebhookSubscriptionRepo
	webhookEventRepo          repo.WebhookEventRepo
	platformsRepo             repo.PlatformRepo
	paymentProviderConfigRepo repo.PaymentProviderConfigRepo
	registry                  providers.ProviderRegistry
	dispatcher                providers.WebhookDispatcher
	logger                    zerolog.Logger
}

// NewService creates a new webhook service. dispatcher is the provider
// webhook event dispatcher used to process normalized events (e.g. the
// HighLevel INSTALL/UNINSTALL lifecycle). It may be nil; when nil, events are
// persisted but not dispatched.
func NewService(
	integrationsRepo repo.IntegrationRepo,
	webhookRepo repo.WebhookSubscriptionRepo,
	webhookEventRepo repo.WebhookEventRepo,
	platformsRepo repo.PlatformRepo,
	paymentProviderConfigRepo repo.PaymentProviderConfigRepo,
	registry providers.ProviderRegistry,
	dispatcher providers.WebhookDispatcher,
	logger zerolog.Logger,
) *Service {
	return &Service{
		integrationsRepo:          integrationsRepo,
		webhookRepo:               webhookRepo,
		webhookEventRepo:          webhookEventRepo,
		platformsRepo:             platformsRepo,
		paymentProviderConfigRepo: paymentProviderConfigRepo,
		registry:                  registry,
		dispatcher:                dispatcher,
		logger:                    logger,
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

	webhookProvider := provider.WebhookProvider()
	if webhookProvider == nil {
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

	webhookProvider := provider.WebhookProvider()
	if webhookProvider == nil {
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

	webhookProvider := provider.WebhookProvider()
	if webhookProvider == nil {
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

	// For HighLevel webhooks, the GHL appId is NOT an RVPay UUID. Resolve the
	// actual RVPay integration UUID deterministically from the GHL locationId.
	//
	// For INSTALL events, the pre-provisioned integration mapping
	// (integration.external_account_id = GHL locationId) is authoritative and
	// must be resolved BEFORE the payment_provider_configs record exists. The
	// INSTALL handler is dispatched first so it can create the
	// payment_provider_configs record idempotently; the webhook subscription
	// requirement is relaxed for INSTALL because a first installation has no
	// subscription yet (it is registered after OAuth completes). For
	// non-INSTALL events, the existing behavior is preserved: resolve via the
	// payment_provider_configs table, require the webhook subscription, then
	// dispatch.
	var integrationID uuid.UUID
	isInstall := providerID == "highlevel" && (event.EventType == "INSTALL" || event.EventType == "integration.installed")
	if providerID == "highlevel" && event.LocationID != "" {
		if isInstall {
			// First-install flow: the integration is pre-provisioned with
			// external_account_id = GHL locationId. Resolve it directly so the
			// INSTALL handler can create the payment_provider_configs record.
			integration, err := s.integrationsRepo.GetByExternalAccountID(ctx, event.LocationID)
			if err == repo.ErrNotFound {
				// Fall back to the config mapping (integration already active).
				config, configErr := s.paymentProviderConfigRepo.GetByLocationID(ctx, event.LocationID)
				if configErr == repo.ErrNotFound {
					return ErrIntegrationNotFound
				}
				if configErr != nil {
					return translateError(configErr)
				}
				integrationID = config.IntegrationID
			} else if err != nil {
				return translateError(err)
			} else {
				integrationID = integration.ID
			}
		} else {
			config, err := s.paymentProviderConfigRepo.GetByLocationID(ctx, event.LocationID)
			if err == repo.ErrNotFound {
				return ErrIntegrationNotFound
			}
			if err != nil {
				return translateError(err)
			}
			integrationID = config.IntegrationID
		}
	} else {
		integrationID, err = uuid.Parse(event.IntegrationID)
		if err != nil {
			return ErrInvalidPayload
		}
	}

	// Idempotency: record the event atomically. The unique constraint on
	// (integration_id, provider_event_id) plus ON CONFLICT DO NOTHING makes
	// duplicate deliveries race-safe: a concurrent retry of the same event
	// will not insert a second row and is treated as a duplicate.
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		s.logger.Error().Err(err).Str("event_id", event.ProviderEventID).Msg("Webhook payload marshal failed")
		return ErrInvalidPayload
	}

	_, err = s.webhookEventRepo.Create(ctx, integrationID, event.ProviderEventID, event.EventType, payload)
	if err == repo.ErrDuplicate {
		s.logger.Info().Str("event_id", event.ProviderEventID).Str("integration_id", integrationID.String()).Msg("Duplicate webhook event ignored")
		return ErrDuplicateEvent
	}
	if err != nil {
		return translateError(err)
	}

	// For INSTALL events, dispatch the INSTALL handler BEFORE the webhook
	// subscription check. The INSTALL handler creates/finds the
	// payment_provider_configs record idempotently. A first installation has
	// no webhook subscription yet (it is registered after OAuth completes), so
	// the INSTALL event must not fail merely because the subscription or the
	// config does not exist yet.
	if isInstall {
		if s.dispatcher != nil {
			err = s.dispatcher.Dispatch(ctx, event)
			if err != nil {
				s.logger.Error().Err(err).Str("event_id", event.ProviderEventID).Msg("Webhook event dispatch failed")
				return err
			}
		}
		s.logger.Info().Str("provider", providerID).Str("event_type", event.EventType).Str("event_id", event.ProviderEventID).Msg("Webhook processed successfully")
		return nil
	}

	webhook, err := s.webhookRepo.GetByIntegrationIDAndEndpoint(ctx, integrationID, "")
	if err == repo.ErrNotFound {
		return ErrWebhookNotFound
	}
	if err != nil {
		return translateError(err)
	}

	s.logger.Info().Str("event_id", event.ProviderEventID).Str("integration_id", integrationID.String()).Msg("Webhook subscription exists, processing event")

	lastDelivery := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	_, err = s.webhookRepo.UpdateLastDelivery(ctx, webhook.ID, lastDelivery)
	if err != nil {
		return translateError(err)
	}

	if s.dispatcher != nil {
		err = s.dispatcher.Dispatch(ctx, event)
		if err != nil {
			s.logger.Error().Err(err).Str("event_id", event.ProviderEventID).Msg("Webhook event dispatch failed")
			return err
		}
	}

	s.logger.Info().Str("provider", providerID).Str("event_type", event.EventType).Str("event_id", event.ProviderEventID).Msg("Webhook processed successfully")

	return nil
}
