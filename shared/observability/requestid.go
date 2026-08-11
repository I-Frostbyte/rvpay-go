// Package observability provides the request-correlation primitives shared
// by RVPay services: a context-based request ID, a gRPC unary logging
// interceptor, and HTTP request-ID/access-log middleware.
//
// Responsibility: carry a request ID through a single service invocation and
// produce structured request lifecycle logs for gRPC and HTTP.
//
// Consumers: service bootstrap code (cmd/grpc-service) for Clients and
// Transactions.
//
// Non-responsibilities: security (no auth), metrics, tracing, and business
// logic. Tracing/metrics have no documented architecture requirement.
package observability

import (
	"context"

	"github.com/google/uuid"
)

// Header is the standard HTTP/gRPC metadata key used for the request ID.
const Header = "X-Request-ID"

type ctxKey struct{}

// NewRequestID returns a new random request ID.
func NewRequestID() string {
	return uuid.NewString()
}

// WithRequestID returns a context carrying id.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// RequestIDFromContext returns the request ID carried by ctx, or "" when
// absent.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

// GetOrCreate returns the request ID from ctx when present, otherwise it
// generates a new one and stores it in a derived context.
func GetOrCreate(ctx context.Context) (context.Context, string) {
	if id := RequestIDFromContext(ctx); id != "" {
		return ctx, id
	}
	id := NewRequestID()
	return WithRequestID(ctx, id), id
}