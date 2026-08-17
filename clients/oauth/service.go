package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/I-Frostbyte/rvpay-go/clients/providers"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// ProviderConfigSettings holds the configuration used to build the HighLevel
// Custom Payment Provider configuration. All values come from environment
// configuration; none are hard-coded.
type ProviderConfigSettings struct {
	// Name is the display name of the payment provider.
	Name string
	// Description is the description of the payment provider.
	Description string
	// ImageURL is the publicly accessible image URL of the payment provider.
	ImageURL string
	// PaymentsURL is the frontend checkout URL supplied to HighLevel.
	PaymentsURL string
	// QueryURL is the backend query URL supplied to HighLevel.
	QueryURL string
}

// Service manages OAuth flows for provider integrations.
type Service struct {
	integrationsRepo repo.IntegrationRepo
	oauthRepo        repo.OAuthTokenRepo
	clientsRepo      repo.ClientRepo
	platformsRepo    repo.PlatformRepo
	oauthStateRepo   repo.OAuthStateRepo
	configRepo       repo.PaymentProviderConfigRepo
	registry         providers.ProviderRegistry
	redirectURI      string
	stateTTL         time.Duration
	providerConfig   ProviderConfigSettings
	logger           zerolog.Logger
}

// NewService creates a new OAuth service. redirectURI is the configured
// callback URL used for the OAuth authorization and token exchange; it must
// come from configuration (HIGHLEVEL_REDIRECT_URI), never be hard-coded.
// oauthStateRepo persists OAuth state so the callback can securely recover
// the client/platform context and resist CSRF/replay attacks.
//
// configRepo persists the HighLevel Custom Payment Provider configuration
// created during the registration lifecycle. It may be nil if the provider
// does not support Custom Payment Provider operations.
//
// providerConfig holds the configuration used to build the provider
// configuration sent to HighLevel. It is only used when the provider supports
// the payment provider capability.
func NewService(
	integrationsRepo repo.IntegrationRepo,
	oauthRepo repo.OAuthTokenRepo,
	clientsRepo repo.ClientRepo,
	platformsRepo repo.PlatformRepo,
	oauthStateRepo repo.OAuthStateRepo,
	configRepo repo.PaymentProviderConfigRepo,
	registry providers.ProviderRegistry,
	redirectURI string,
	providerConfig ProviderConfigSettings,
	logger zerolog.Logger,
) *Service {
	return &Service{
		integrationsRepo: integrationsRepo,
		oauthRepo:        oauthRepo,
		clientsRepo:      clientsRepo,
		platformsRepo:    platformsRepo,
		oauthStateRepo:   oauthStateRepo,
		configRepo:       configRepo,
		registry:         registry,
		redirectURI:      redirectURI,
		stateTTL:         10 * time.Minute,
		providerConfig:   providerConfig,
		logger:           logger,
	}
}

// AuthorizationURL generates the OAuth authorization URL for a client and platform.
func (s *Service) AuthorizationURL(ctx context.Context, clientID, platformID uuid.UUID, state string) (string, error) {
	platform, err := s.platformsRepo.GetByID(ctx, platformID)
	if err == repo.ErrNotFound {
		return "", ErrPlatformNotFound
	}
	if err != nil {
		return "", translateError(err)
	}

	if !platform.Enabled {
		return "", ErrPlatformDisabled
	}

	provider, ok := s.registry.Get(platform.Slug)
	if !ok {
		return "", ErrProviderNotSupported
	}

	authURL, err := provider.OAuthProvider().GenerateAuthorizationURL(ctx, state, s.redirectURI)
	if err != nil {
		return "", err
	}

	s.logger.Info().Str("client_id", clientID.String()).Str("platform_id", platformID.String()).Str("provider", provider.ID()).Msg("OAuth authorization URL generated")

	return authURL, nil
}

// BeginAuthorization starts an OAuth authorization flow for a client and
// platform. It generates a cryptographically random state, persists it with
// the client/platform context and an expiry, and returns the provider
// authorization URL. The returned state must be validated by HandleCallback
// when the provider redirects the user back.
func (s *Service) BeginAuthorization(ctx context.Context, clientID, platformID uuid.UUID) (authURL, state string, err error) {
	platform, err := s.platformsRepo.GetByID(ctx, platformID)
	if err == repo.ErrNotFound {
		return "", "", ErrPlatformNotFound
	}
	if err != nil {
		return "", "", translateError(err)
	}

	if !platform.Enabled {
		return "", "", ErrPlatformDisabled
	}

	client, err := s.clientsRepo.GetByID(ctx, clientID)
	if err == repo.ErrNotFound {
		return "", "", ErrClientNotFound
	}
	if err != nil {
		return "", "", translateError(err)
	}

	if client.Status != sqlc.ClientStatusACTIVE {
		return "", "", ErrClientInactive
	}

	provider, ok := s.registry.Get(platform.Slug)
	if !ok {
		return "", "", ErrProviderNotSupported
	}

	state, err = providers.GenerateState()
	if err != nil {
		return "", "", err
	}

	_, err = s.oauthStateRepo.Create(ctx, state, clientID, platformID, time.Now().Add(s.stateTTL))
	if err != nil {
		return "", "", translateError(err)
	}

	authURL, err = provider.OAuthProvider().GenerateAuthorizationURL(ctx, state, s.redirectURI)
	if err != nil {
		return "", "", err
	}

	s.logger.Info().Str("client_id", clientID.String()).Str("platform_id", platformID.String()).Str("provider", provider.ID()).Msg("OAuth authorization flow started")

	return authURL, state, nil
}

// HandleCallback processes an OAuth callback from a provider.
//
// The `code` is always required. The `state` is optional:
//
//   - When `state` is present, the existing state-based flow is preserved: the
//     state is validated (must exist, be unexpired, and be unconsumed),
//     atomically consumed to prevent replay, and used to recover the
//     client/platform context.
//   - When `state` is absent (HighLevel Marketplace OAuth does not return a
//     state), the authorization code is exchanged first to obtain the GHL
//     locationId. The locationId is then resolved to the integration via the
//     deterministic mapping: first by `integration.external_account_id` (the
//     GHL location identifier established during provisioning/activation), then
//     by `payment_provider_configs.location_id`. The integration's
//     client_id/platform_id are used to continue the existing
//     ProcessCallback/integration flow.
func (s *Service) HandleCallback(ctx context.Context, code, state string) (*CallbackResult, error) {
	if code == "" {
		return nil, ErrMissingCode
	}

	if state != "" {
		// Atomically consume the state. ConsumeOAuthState only succeeds when the
		// state exists, is not already consumed, and has not expired. This both
		// validates the state and prevents replay attacks in a single operation.
		record, err := s.oauthStateRepo.Consume(ctx, state)
		if err == repo.ErrNotFound {
			// Distinguish expired/consumed from unknown for clearer errors.
			existing, getErr := s.oauthStateRepo.GetByState(ctx, state)
			if getErr == nil {
				if existing.ConsumedAt.Valid {
					return nil, ErrStateConsumed
				}
				return nil, ErrStateExpired
			}
			return nil, ErrInvalidState
		}
		if err != nil {
			return nil, translateError(err)
		}

		return s.ProcessCallback(ctx, record.ClientID, record.PlatformID, code, state)
	}

	// No state: resolve the client/platform context from the GHL locationId.
	// Exchange the authorization code first to obtain the locationId, then
	// resolve locationId -> integration via the deterministic mapping.
	if s.configRepo == nil {
		return nil, ErrProviderConfigRepoNotConfigured
	}

	provider, ok := s.registry.Get("highlevel")
	if !ok {
		return nil, ErrProviderNotSupported
	}

	tokenResp, err := provider.OAuthProvider().ExchangeCode(ctx, code, s.redirectURI)
	if err != nil {
		s.logger.Error().Err(err).Msg("OAuth token exchange failed during stateless callback")
		return nil, ErrTokenExchangeFailed
	}
	if tokenResp.LocationID == "" {
		return nil, ErrMissingLocationID
	}

	// Resolve the integration deterministically from the GHL locationId.
	// First try the provisioning mapping: integration.external_account_id =
	// GHL locationId. This is the authoritative mapping established when the
	// integration is activated. If that is not set, fall back to the
	// payment_provider_configs.location_id mapping (created during provider
	// registration). If neither resolves, fail clearly rather than selecting
	// an arbitrary integration or client.
	integration, err := s.integrationsRepo.GetByExternalAccountID(ctx, tokenResp.LocationID)
	if err == repo.ErrNotFound {
		// Fall back to the payment provider config mapping.
		config, configErr := s.configRepo.GetByLocationID(ctx, tokenResp.LocationID)
		if configErr == repo.ErrNotFound {
			return nil, ErrIntegrationNotFound
		}
		if configErr != nil {
			return nil, translateError(configErr)
		}
		integration, err = s.integrationsRepo.GetByID(ctx, config.IntegrationID)
		if err == repo.ErrNotFound {
			return nil, ErrIntegrationNotFound
		}
		if err != nil {
			return nil, translateError(err)
		}
	} else if err != nil {
		return nil, translateError(err)
	}

	return s.ProcessCallback(ctx, integration.ClientID, integration.PlatformID, code, state)
}

// CallbackResult represents the result of an OAuth callback.
type CallbackResult struct {
	IntegrationID  uuid.UUID
	ClientID       uuid.UUID
	PlatformID     uuid.UUID
	AccessToken    string
	RefreshToken   string
	ExpiresAt      time.Time
	Scope          string
	ProviderUserID string
	// ProviderRegistered indicates whether the HighLevel Custom Payment
	// Provider registration completed successfully during the OAuth callback.
	// It is false when the provider does not support payment provider
	// operations or when registration failed (in which case the integration
	// remains installed and RegisterProvider can be retried).
	ProviderRegistered bool
	// ProviderRegistrationError is set when provider registration failed but
	// the OAuth installation itself succeeded. The integration remains
	// installed; the error is informational so the caller can decide whether
	// to surface it or retry registration.
	ProviderRegistrationError error
}

// ProcessCallback processes the OAuth callback and creates the integration.
// After the integration and OAuth token are persisted, it triggers the
// HighLevel Custom Payment Provider registration lifecycle (provider
// association and configuration) when the provider supports it.
//
// If an integration already exists for the client/platform with status
// CREATED (a pre-provisioned integration awaiting OAuth completion), it is
// reused and activated rather than returning ErrIntegrationAlreadyExists.
// This supports the HighLevel Marketplace first-install flow where the
// integration is pre-created before the OAuth callback arrives.
func (s *Service) ProcessCallback(ctx context.Context, clientID, platformID uuid.UUID, code, state string) (*CallbackResult, error) {
	platform, err := s.platformsRepo.GetByID(ctx, platformID)
	if err == repo.ErrNotFound {
		return nil, ErrPlatformNotFound
	}
	if err != nil {
		return nil, translateError(err)
	}

	if !platform.Enabled {
		return nil, ErrPlatformDisabled
	}

	provider, ok := s.registry.Get(platform.Slug)
	if !ok {
		return nil, ErrProviderNotSupported
	}

	client, err := s.clientsRepo.GetByID(ctx, clientID)
	if err == repo.ErrNotFound {
		return nil, ErrClientNotFound
	}
	if err != nil {
		return nil, translateError(err)
	}

	if client.Status != sqlc.ClientStatusACTIVE {
		return nil, ErrClientInactive
	}

	tokenResp, err := provider.OAuthProvider().ExchangeCode(ctx, code, s.redirectURI)
	if err != nil {
		s.logger.Error().Err(err).Str("client_id", clientID.String()).Str("platform_id", platformID.String()).Msg("OAuth token exchange failed")
		return nil, ErrTokenExchangeFailed
	}

	providerUserID, err := provider.OAuthProvider().GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		s.logger.Error().Err(err).Str("client_id", clientID.String()).Str("platform_id", platformID.String()).Msg("OAuth user info retrieval failed")
		return nil, ErrUserInfoFailed
	}

	// Determine the integration to use. If an integration already exists for
	// this client/platform:
	//   - CREATED: reuse it (pre-provisioned integration awaiting OAuth
	//     completion). Activate it and continue the token/registration flow.
	//   - otherwise: return ErrIntegrationAlreadyExists (genuine conflict).
	var integration sqlc.Integration
	existing, err := s.integrationsRepo.GetByClientAndPlatform(ctx, clientID, platformID)
	if err == nil {
		if existing.Status != sqlc.IntegrationStatusCREATED {
			return nil, ErrIntegrationAlreadyExists
		}
		// Reuse the pre-provisioned CREATED integration. Activate it. The
		// external_account_id is set to the GHL locationId when the provider
		// registration persists the payment_provider_configs record; the
		// locationId is the deterministic GHL sub-account identifier.
		integration, err = s.integrationsRepo.UpdateStatus(ctx, existing.ID, sqlc.IntegrationStatusACTIVE)
		if err != nil {
			return nil, translateError(err)
		}
		s.logger.Info().Str("integration_id", integration.ID.String()).Str("client_id", clientID.String()).Str("platform_id", platformID.String()).Msg("reused pre-provisioned CREATED integration")
	} else if !errors.Is(err, repo.ErrNotFound) {
		return nil, translateError(err)
	} else {
		integration, err = s.integrationsRepo.Create(ctx, clientID, platformID, providerUserID, sqlc.IntegrationStatusACTIVE)
		if err != nil {
			return nil, translateError(err)
		}
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	_, err = s.oauthRepo.Create(ctx, integration.ID, tokenResp.AccessToken, tokenResp.RefreshToken, expiresAt, tokenResp.Scope, tokenResp.TokenType)
	if err != nil {
		s.logger.Error().Err(err).Str("integration_id", integration.ID.String()).Msg("OAuth token persistence failed")
		return nil, translateError(err)
	}

	result := &CallbackResult{
		IntegrationID:  integration.ID,
		ClientID:       clientID,
		PlatformID:     platformID,
		AccessToken:    tokenResp.AccessToken,
		RefreshToken:   tokenResp.RefreshToken,
		ExpiresAt:      expiresAt,
		Scope:          tokenResp.Scope,
		ProviderUserID: providerUserID,
	}

	// Trigger the HighLevel Custom Payment Provider registration lifecycle.
	// The OAuth installation is already persisted; registration is a
	// best-effort follow-up that must not roll back the installation. The
	// location ID used for registration is the actual HighLevel locationId
	// from the OAuth token response, not the HighLevel user ID.
	if provider.HasCapability(providers.CapabilityPaymentProvider) {
		regErr := s.RegisterProvider(ctx, integration.ID, tokenResp.LocationID, tokenResp.AccessToken)
		if regErr != nil {
			s.logger.Warn().Err(regErr).Str("integration_id", integration.ID.String()).Str("location_id", tokenResp.LocationID).Msg("HighLevel provider registration failed; integration remains installed")
			result.ProviderRegistrationError = regErr
		} else {
			result.ProviderRegistered = true
		}
	}

	s.logger.Info().Str("integration_id", integration.ID.String()).Str("client_id", clientID.String()).Str("platform_id", platformID.String()).Str("provider_user_id", providerUserID).Bool("provider_registered", result.ProviderRegistered).Msg("OAuth callback processed successfully")

	return result, nil
}

// RegisterProvider performs the HighLevel Custom Payment Provider registration
// lifecycle for an installed integration. It:
//
//  1. Creates the provider association (POST /payments/custom-provider/provider).
//  2. Creates the provider configuration (POST /payments/custom-provider/connect).
//  3. Persists the provider configuration locally.
//
// The operation is idempotent: if the provider is already associated or
// configured, the existing configuration is fetched and persisted instead of
// creating a duplicate. Registration failures return a typed error; the
// integration remains installed and the operation can be retried safely.
func (s *Service) RegisterProvider(ctx context.Context, integrationID uuid.UUID, locationID, accessToken string) error {
	if s.configRepo == nil {
		return ErrProviderConfigRepoNotConfigured
	}
	if locationID == "" {
		return ErrMissingLocationID
	}
	if accessToken == "" {
		return ErrMissingAccessToken
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

	paymentClient := provider.PaymentProvider()
	if paymentClient == nil {
		return ErrPaymentProviderNotSupported
	}

	// Step 1: Create the provider association. If the provider is already
	// associated, HighLevel may return a 400/422; we treat that as idempotent
	// and continue to the configuration step.
	err = paymentClient.CreateProviderAssociation(ctx, accessToken, locationID)
	if err != nil {
		if errors.Is(err, providers.ErrBadRequest) || errors.Is(err, providers.ErrUnprocessableEntity) {
			s.logger.Info().Str("integration_id", integrationID.String()).Str("location_id", locationID).Msg("provider association already exists; continuing with configuration")
		} else {
			return ErrProviderAssociationFailed
		}
	}

	// Step 2: Create the provider configuration. If the configuration already
	// exists, fetch the existing configuration instead of creating a duplicate.
	config := providers.ProviderConfig{
		Name:                         s.providerConfig.Name,
		Description:                  s.providerConfig.Description,
		ImageURL:                     s.providerConfig.ImageURL,
		LocationID:                   locationID,
		QueryURL:                     s.providerConfig.QueryURL,
		PaymentsURL:                  s.providerConfig.PaymentsURL,
		SupportsSubscriptionSchedule: false, // RVPay supports one-time payments only.
	}

	err = paymentClient.CreateProviderConfig(ctx, accessToken, config)
	if err != nil {
		if errors.Is(err, providers.ErrBadRequest) || errors.Is(err, providers.ErrUnprocessableEntity) {
			// The configuration may already exist. Fetch the existing
			// configuration to confirm and persist it locally.
			s.logger.Info().Str("integration_id", integrationID.String()).Str("location_id", locationID).Msg("provider config may already exist; fetching existing configuration")
			existing, fetchErr := paymentClient.FetchProviderConfig(ctx, accessToken, locationID)
			if fetchErr != nil {
				return ErrProviderConfigFailed
			}
			config = *existing
		} else {
			return ErrProviderConfigFailed
		}
	}

	// Step 3: Persist the provider configuration locally. The provider API key
	// is a generated random value used to authenticate HighLevel query
	// requests; it is distinct from the OAuth access token and the pawaPay
	// API key.
	apiKey, err := generateAPIKey()
	if err != nil {
		return ErrAPIKeyGenerationFailed
	}

	_, err = s.configRepo.Create(ctx, integrationID, config.Name, config.Description, config.ImageURL, config.LocationID, config.QueryURL, config.PaymentsURL, config.SupportsSubscriptionSchedule, apiKey)
	if err == repo.ErrDuplicate {
		// The config already exists locally; update it instead.
		_, err = s.configRepo.Update(ctx, integrationID, config.Name, config.Description, config.ImageURL, config.LocationID, config.QueryURL, config.PaymentsURL, config.SupportsSubscriptionSchedule, apiKey)
		if err != nil {
			return translateError(err)
		}
	} else if err != nil {
		return translateError(err)
	}

	s.logger.Info().Str("integration_id", integrationID.String()).Str("location_id", locationID).Msg("HighLevel provider registration completed")

	return nil
}

// RefreshAccessToken refreshes an OAuth access token for an integration.
func (s *Service) RefreshAccessToken(ctx context.Context, integrationID uuid.UUID) error {
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

	oauthToken, err := s.oauthRepo.GetByIntegrationID(ctx, integrationID)
	if err == repo.ErrNotFound {
		return ErrOAuthTokenNotFound
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

	tokenResp, err := provider.OAuthProvider().RefreshToken(ctx, oauthToken.RefreshToken)
	if err != nil {
		s.logger.Error().Err(err).Str("integration_id", integrationID.String()).Msg("OAuth token refresh failed")
		return ErrTokenRefreshFailed
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	_, err = s.oauthRepo.Update(ctx, oauthToken.ID, tokenResp.AccessToken, tokenResp.RefreshToken, expiresAt, tokenResp.Scope, tokenResp.TokenType)
	if err != nil {
		return translateError(err)
	}

	s.logger.Info().Str("integration_id", integrationID.String()).Msg("OAuth token refreshed")

	return nil
}

// ValidateToken validates an OAuth access token for an integration.
func (s *Service) ValidateToken(ctx context.Context, integrationID uuid.UUID) (bool, error) {
	oauthToken, err := s.oauthRepo.GetByIntegrationID(ctx, integrationID)
	if err == repo.ErrNotFound {
		return false, ErrOAuthTokenNotFound
	}
	if err != nil {
		return false, translateError(err)
	}

	if time.Now().After(oauthToken.ExpiresAt) {
		return false, nil
	}

	integration, err := s.integrationsRepo.GetByID(ctx, integrationID)
	if err == repo.ErrNotFound {
		return false, ErrIntegrationNotFound
	}
	if err != nil {
		return false, translateError(err)
	}

	if integration.Status != sqlc.IntegrationStatusACTIVE {
		return false, ErrIntegrationNotActive
	}

	platform, err := s.platformsRepo.GetByID(ctx, integration.PlatformID)
	if err == repo.ErrNotFound {
		return false, ErrPlatformNotFound
	}
	if err != nil {
		return false, translateError(err)
	}

	provider, ok := s.registry.Get(platform.Slug)
	if !ok {
		return false, ErrProviderNotSupported
	}

	valid, err := provider.OAuthProvider().ValidateToken(ctx, oauthToken.AccessToken)
	if err != nil {
		return false, err
	}

	return valid, nil
}

// generateAPIKey generates a cryptographically random API key used to
// authenticate HighLevel payment query requests. It is distinct from the
// OAuth access token and the pawaPay API key.
func generateAPIKey() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
