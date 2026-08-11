package webhooks

import (
	"errors"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrWebhookNotFound is returned when a webhook subscription does not exist.
	ErrWebhookNotFound = status.Error(codes.NotFound, "webhook not found")
	// ErrWebhookAlreadyExists is returned when a webhook subscription already exists.
	ErrWebhookAlreadyExists = status.Error(codes.AlreadyExists, "webhook already exists")
	// ErrInvalidSignature is returned when webhook signature validation fails.
	ErrInvalidSignature = status.Error(codes.InvalidArgument, "invalid webhook signature")
	// ErrInvalidPayload is returned when webhook payload parsing fails.
	ErrInvalidPayload = status.Error(codes.InvalidArgument, "invalid webhook payload")
	// ErrDuplicateEvent is returned when a duplicate webhook event is detected.
	ErrDuplicateEvent = status.Error(codes.AlreadyExists, "duplicate webhook event")
	// ErrProviderNotSupported is returned when a provider is not supported.
	ErrProviderNotSupported = status.Error(codes.FailedPrecondition, "provider not supported")
	// ErrIntegrationNotFound is returned when an integration does not exist.
	ErrIntegrationNotFound = status.Error(codes.NotFound, "integration not found")
	// ErrIntegrationNotActive is returned when an integration is not active.
	ErrIntegrationNotActive = status.Error(codes.FailedPrecondition, "integration is not active")
	// ErrPlatformNotFound is returned when a platform does not exist.
	ErrPlatformNotFound = status.Error(codes.NotFound, "platform not found")
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
