package oauth

import (
	"errors"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrPlatformNotFound is returned when a platform does not exist.
	ErrPlatformNotFound = status.Error(codes.NotFound, "platform not found")
	// ErrPlatformDisabled is returned when a disabled platform cannot be used for OAuth.
	ErrPlatformDisabled = status.Error(codes.FailedPrecondition, "platform is disabled")
	// ErrProviderNotSupported is returned when a provider is not registered.
	ErrProviderNotSupported = status.Error(codes.FailedPrecondition, "provider not supported")
	// ErrClientNotFound is returned when a client does not exist.
	ErrClientNotFound = status.Error(codes.NotFound, "client not found")
	// ErrClientInactive is returned when an inactive client cannot perform OAuth.
	ErrClientInactive = status.Error(codes.FailedPrecondition, "client is not active")
	// ErrTokenExchangeFailed is returned when the authorization code exchange fails.
	ErrTokenExchangeFailed = status.Error(codes.Internal, "token exchange failed")
	// ErrUserInfoFailed is returned when retrieving user info from the provider fails.
	ErrUserInfoFailed = status.Error(codes.Internal, "user info retrieval failed")
	// ErrIntegrationAlreadyExists is returned when an integration already exists for the client and platform.
	ErrIntegrationAlreadyExists = status.Error(codes.AlreadyExists, "integration already exists")
	// ErrIntegrationNotFound is returned when an integration does not exist.
	ErrIntegrationNotFound = status.Error(codes.NotFound, "integration not found")
	// ErrIntegrationNotActive is returned when an integration is not active.
	ErrIntegrationNotActive = status.Error(codes.FailedPrecondition, "integration is not active")
	// ErrOAuthTokenNotFound is returned when OAuth tokens are not found for an integration.
	ErrOAuthTokenNotFound = status.Error(codes.NotFound, "OAuth token not found")
	// ErrTokenRefreshFailed is returned when refreshing an access token fails.
	ErrTokenRefreshFailed = status.Error(codes.Internal, "token refresh failed")
	// ErrInvalidState is returned when the OAuth state is missing, unknown,
	// expired, or already consumed.
	ErrInvalidState = status.Error(codes.InvalidArgument, "invalid OAuth state")
	// ErrStateExpired is returned when the OAuth state has expired.
	ErrStateExpired = status.Error(codes.InvalidArgument, "OAuth state expired")
	// ErrStateConsumed is returned when the OAuth state was already used.
	ErrStateConsumed = status.Error(codes.InvalidArgument, "OAuth state already used")
	// ErrMissingCode is returned when the OAuth callback is missing the code.
	ErrMissingCode = status.Error(codes.InvalidArgument, "authorization code is required")
	// ErrMissingState is returned when the OAuth callback is missing the state.
	ErrMissingState = status.Error(codes.InvalidArgument, "OAuth state is required")

	// ErrProviderConfigRepoNotConfigured is returned when the payment provider
	// config repository is not configured on the OAuth service.
	ErrProviderConfigRepoNotConfigured = status.Error(codes.FailedPrecondition, "payment provider config repository not configured")
	// ErrMissingLocationID is returned when no location ID is provided for
	// provider registration.
	ErrMissingLocationID = status.Error(codes.InvalidArgument, "location ID is required")
	// ErrMissingAccessToken is returned when no access token is provided for
	// provider registration.
	ErrMissingAccessToken = status.Error(codes.InvalidArgument, "access token is required")
	// ErrPaymentProviderNotSupported is returned when the provider does not
	// support Custom Payment Provider operations.
	ErrPaymentProviderNotSupported = status.Error(codes.FailedPrecondition, "payment provider not supported")
	// ErrProviderAssociationFailed is returned when creating the provider
	// association with HighLevel fails.
	ErrProviderAssociationFailed = status.Error(codes.Internal, "provider association failed")
	// ErrProviderConfigFailed is returned when creating or fetching the
	// provider configuration with HighLevel fails.
	ErrProviderConfigFailed = status.Error(codes.Internal, "provider configuration failed")
	// ErrAPIKeyGenerationFailed is returned when generating the provider API
	// key fails.
	ErrAPIKeyGenerationFailed = status.Error(codes.Internal, "provider API key generation failed")
)

// translateError converts repository errors to business errors.
func translateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, repo.ErrNotFound) {
		return status.Error(codes.NotFound, "record not found")
	}
	if errors.Is(err, repo.ErrDuplicate) {
		return status.Error(codes.AlreadyExists, "record already exists")
	}
	if errors.Is(err, repo.ErrConstraint) {
		return status.Error(codes.FailedPrecondition, "constraint violation")
	}
	return status.Error(codes.Internal, "internal error")
}
