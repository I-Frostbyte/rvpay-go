package providers

import (
	"context"
)

// Provider represents a marketplace platform provider.
type Provider interface {
	// ID returns the unique provider identifier.
	ID() string
	// Name returns the human-readable provider name.
	Name() string
	// OAuthProvider returns the OAuth implementation for this provider.
	OAuthProvider() OAuthProvider
}

// OAuthProvider defines the OAuth lifecycle for a provider.
type OAuthProvider interface {
	// GenerateAuthorizationURL creates the URL to redirect the user for authorization.
	GenerateAuthorizationURL(ctx context.Context, state string, redirectURI string) (string, error)
	// ExchangeCode exchanges an authorization code for access and refresh tokens.
	ExchangeCode(ctx context.Context, code string, redirectURI string) (*TokenResponse, error)
	// RefreshToken refreshes an access token using a refresh token.
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
	// GetUserInfo retrieves the provider account identifier for the authenticated user.
	GetUserInfo(ctx context.Context, accessToken string) (string, error)
	// ValidateToken checks if an access token is still valid.
	ValidateToken(ctx context.Context, accessToken string) (bool, error)
}

// TokenResponse represents the OAuth token response from a provider.
type TokenResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	TokenType    string
	Scope        string
}

// ProviderRegistry manages available providers.
type ProviderRegistry interface {
	// Register adds a provider to the registry.
	Register(provider Provider)
	// Get returns a provider by ID.
	Get(id string) (Provider, bool)
	// List returns all registered providers.
	List() []Provider
}