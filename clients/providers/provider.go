package providers

import "context"

// Capability represents a provider capability.
type Capability string

const (
	CapabilityOAuth           Capability = "oauth"
	CapabilityWebhooks        Capability = "webhooks"
	CapabilityTokenRefresh    Capability = "token_refresh"
	CapabilityInstallation    Capability = "installation"
	CapabilityUninstallation  Capability = "uninstallation"
	CapabilityHealthCheck     Capability = "health_check"
	CapabilityPaymentProvider Capability = "payment_provider"
)

// Provider represents a marketplace platform provider.
type Provider interface {
	// ID returns the unique provider identifier.
	ID() string
	// Name returns the human-readable provider name.
	Name() string
	// Capabilities returns the list of capabilities this provider supports.
	Capabilities() []Capability
	// HasCapability checks if the provider supports a specific capability.
	HasCapability(capability Capability) bool
	// OAuthProvider returns the OAuth implementation for this provider.
	OAuthProvider() OAuthProvider
	// WebhookProvider returns the webhook implementation for this provider.
	WebhookProvider() WebhookProvider
	// PaymentProvider returns the Custom Payment Provider client for this
	// provider, or nil if the provider does not support it.
	PaymentProvider() PaymentProviderClient
}

// PaymentProviderCapable is implemented by providers that support Custom
// Payment Provider operations.
type PaymentProviderCapable interface {
	// PaymentProvider returns the Custom Payment Provider client
	// implementation for this provider.
	PaymentProvider() PaymentProviderClient
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
	// GetByCapability returns all providers that support a specific capability.
	GetByCapability(capability Capability) []Provider
}
