package providers

import "context"

// PaymentProviderClient defines the outbound HighLevel Custom Payment Provider
// API operations. These are outbound integrations performed by the Clients
// service on behalf of an installed location; they are NOT exposed as RVPay
// public endpoints.
//
// Each operation requires the installed location's OAuth access token. The
// access token is used only to authenticate the outbound call and is never
// logged or returned in errors.
type PaymentProviderClient interface {
	// CreateProviderAssociation creates the association between the
	// Marketplace app and the HighLevel location.
	//
	// POST /payments/custom-provider/provider
	CreateProviderAssociation(ctx context.Context, accessToken, locationID string) error

	// CreateProviderConfig creates the provider configuration for a location.
	//
	// POST /payments/custom-provider/connect
	CreateProviderConfig(ctx context.Context, accessToken string, config ProviderConfig) error

	// FetchProviderConfig fetches the provider configuration for a location.
	//
	// GET /payments/custom-provider/connect
	FetchProviderConfig(ctx context.Context, accessToken, locationID string) (*ProviderConfig, error)

	// DisconnectProvider disconnects the provider configuration for a location.
	//
	// DELETE /payments/custom-provider/connect
	DisconnectProvider(ctx context.Context, accessToken, locationID string) error
}

// ProviderConfig is the provider configuration sent to HighLevel. It is built
// from RVPay configuration and the correct location; it is never hard-coded.
type ProviderConfig struct {
	// Name is the display name of the payment provider.
	Name string
	// Description is the description of the payment provider.
	Description string
	// ImageURL is the publicly accessible image URL of the payment provider.
	ImageURL string
	// LocationID is the HighLevel location ID for the installed account.
	LocationID string
	// QueryURL is the backend query URL supplied to HighLevel.
	QueryURL string
	// PaymentsURL is the frontend checkout URL supplied to HighLevel.
	PaymentsURL string
	// SupportsSubscriptionSchedule indicates whether the provider supports
	// subscription scheduling. RVPay currently supports one-time payments
	// only, so this is always false.
	SupportsSubscriptionSchedule bool
}
