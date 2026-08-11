package service

import (
	"errors"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrClientNotFound is returned when a client does not exist.
	ErrClientNotFound = status.Error(codes.NotFound, "client not found")
	// ErrClientAlreadyExists is returned when a client with the same name already exists.
	ErrClientAlreadyExists = status.Error(codes.AlreadyExists, "client already exists")
	// ErrClientInactive is returned when an inactive client attempts an operation.
	ErrClientInactive = status.Error(codes.FailedPrecondition, "client is not active")
	// ErrClientHasIntegrations is returned when a client cannot be deleted due to existing integrations.
	ErrClientHasIntegrations = status.Error(codes.FailedPrecondition, "client has active integrations")
	// ErrPlatformNotFound is returned when a platform does not exist.
	ErrPlatformNotFound = status.Error(codes.NotFound, "platform not found")
	// ErrPlatformDisabled is returned when a disabled platform cannot receive new integrations.
	ErrPlatformDisabled = status.Error(codes.FailedPrecondition, "platform is disabled")
	// ErrPlatformSlugExists is returned when a platform with the same slug already exists.
	ErrPlatformSlugExists = status.Error(codes.AlreadyExists, "platform with this slug already exists")
	// ErrIntegrationNotFound is returned when an integration does not exist.
	ErrIntegrationNotFound = status.Error(codes.NotFound, "integration not found")
	// ErrIntegrationAlreadyExists is returned when a client already has an integration with the same platform.
	ErrIntegrationAlreadyExists = status.Error(codes.AlreadyExists, "integration already exists for this client and platform")
	// ErrIntegrationNotActive is returned when an integration is not in active status.
	ErrIntegrationNotActive = status.Error(codes.FailedPrecondition, "integration is not active")
	// ErrOAuthNotSupported is returned when a platform does not support OAuth.
	ErrOAuthNotSupported = status.Error(codes.FailedPrecondition, "platform does not support OAuth")
	// ErrWebhookNotSupported is returned when a platform does not support webhooks.
	ErrWebhookNotSupported = status.Error(codes.FailedPrecondition, "platform does not support webhooks")
	// ErrWebhookSubscriptionExists is returned when a webhook subscription already exists.
	ErrWebhookSubscriptionExists = status.Error(codes.AlreadyExists, "webhook subscription already exists")
	// ErrWebhookSubscriptionNotFound is returned when a webhook subscription does not exist.
	ErrWebhookSubscriptionNotFound = status.Error(codes.NotFound, "webhook subscription not found")
)

// translateRepoError converts repository errors to business errors.
func translateRepoError(err error) error {
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
