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