package payments

import (
	"errors"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrProviderConfigNotFound is returned when a payment provider
	// configuration does not exist for an integration.
	ErrProviderConfigNotFound = status.Error(codes.NotFound, "payment provider configuration not found")
	// ErrInvalidAPIKey is returned when the provider API key is invalid.
	ErrInvalidAPIKey = status.Error(codes.Unauthenticated, "invalid provider API key")
	// ErrMissingAPIKey is returned when the provider API key is missing.
	ErrMissingAPIKey = status.Error(codes.InvalidArgument, "provider API key is required")
	// ErrMissingType is returned when the query operation type is missing.
	ErrMissingType = status.Error(codes.InvalidArgument, "query type is required")
	// ErrUnsupportedType is returned when the query operation type is unsupported.
	ErrUnsupportedType = status.Error(codes.InvalidArgument, "unsupported query type")
	// ErrMissingTransactionID is returned when both the transaction ID and
	// charge ID are missing.
	ErrMissingTransactionID = status.Error(codes.InvalidArgument, "transaction_id or charge_id is required")
	// ErrTransactionNotFound is returned when the referenced transaction is not found.
	ErrTransactionNotFound = status.Error(codes.NotFound, "transaction not found")
	// ErrInvalidPayload is returned when the webhook payload is malformed.
	ErrInvalidPayload = status.Error(codes.InvalidArgument, "invalid webhook payload")
	// ErrDuplicateEvent is returned when a duplicate webhook event is detected.
	ErrDuplicateEvent = status.Error(codes.AlreadyExists, "duplicate webhook event")
	// ErrIntegrationNotFound is returned when an integration does not exist.
	ErrIntegrationNotFound = status.Error(codes.NotFound, "integration not found")
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
