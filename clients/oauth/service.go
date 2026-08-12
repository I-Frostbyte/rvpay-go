package oauth

import (
	"context"
	"errors"
	"time"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	"github.com/I-Frostbyte/rvpay-go/clients/providers"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Service manages OAuth flows for provider integrations.
type Service struct {
	integrationsRepo repo.IntegrationRepo
	oauthRepo        repo.OAuthTokenRepo
	clientsRepo      repo.ClientRepo
	platformsRepo    repo.PlatformRepo
	oauthStateRepo   repo.OAuthStateRepo
	registry         providers.ProviderRegistry
	redirectURI      string
	stateTTL         time.Duration
	logger           zerolog.Logger
}

// NewService creates a new OAuth service. redirectURI is the configured
// callback URL used for the OAuth authorization and token exchange; it must
// come from configuration (HIGHLEVEL_REDIRECT_URI), never be hard-coded.
// oauthStateRepo persists OAuth state so the callback can securely recover
// the client/platform context and resist CSRF/replay attacks.
func NewService(
	integrationsRepo repo.IntegrationRepo,
	oauthRepo repo.OAuthTokenRepo,
	clientsRepo repo.ClientRepo,
	platformsRepo repo.PlatformRepo,
	oauthStateRepo repo.OAuthStateRepo,
	registry providers.ProviderRegistry,
	redirectURI string,
	logger zerolog.Logger,
) *Service {
	return &Service{
		integrationsRepo: integrationsRepo,
		oauthRepo:        oauthRepo,
		clientsRepo:      clientsRepo,
		platformsRepo:    platformsRepo,
		oauthStateRepo:   oauthStateRepo,
		registry:         registry,
		redirectURI:      redirectURI,
		stateTTL:         10 * time.Minute,
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

// HandleCallback processes an OAuth callback from a provider. It validates the
// state (must exist, be unexpired, and be unconsumed), atomically consumes it
// to prevent replay, recovers the client/platform context, and then completes
// the token exchange and integration creation.
func (s *Service) HandleCallback(ctx context.Context, code, state string) (*CallbackResult, error) {
	if code == "" {
		return nil, ErrMissingCode
	}
	if state == "" {
		return nil, ErrMissingState
	}

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
}

// ProcessCallback processes the OAuth callback and creates the integration.
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

	_, err = s.integrationsRepo.GetByClientAndPlatform(ctx, clientID, platformID)
	if err == nil {
		return nil, ErrIntegrationAlreadyExists
	}
	if !errors.Is(err, repo.ErrNotFound) {
		return nil, translateError(err)
	}

	integration, err := s.integrationsRepo.Create(ctx, clientID, platformID, providerUserID, sqlc.IntegrationStatusACTIVE)
	if err != nil {
		return nil, translateError(err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	_, err = s.oauthRepo.Create(ctx, integration.ID, tokenResp.AccessToken, tokenResp.RefreshToken, expiresAt, tokenResp.Scope, tokenResp.TokenType)
	if err != nil {
		s.logger.Error().Err(err).Str("integration_id", integration.ID.String()).Msg("OAuth token persistence failed")
		return nil, translateError(err)
	}

	s.logger.Info().Str("integration_id", integration.ID.String()).Str("client_id", clientID.String()).Str("platform_id", platformID.String()).Str("provider_user_id", providerUserID).Msg("OAuth callback processed successfully")

	return &CallbackResult{
		IntegrationID:  integration.ID,
		ClientID:       clientID,
		PlatformID:     platformID,
		AccessToken:    tokenResp.AccessToken,
		RefreshToken:   tokenResp.RefreshToken,
		ExpiresAt:      expiresAt,
		Scope:          tokenResp.Scope,
		ProviderUserID: providerUserID,
	}, nil
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
