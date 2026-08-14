package providers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HighLevelProvider implements the unified Provider interface for HighLevel.
type HighLevelProvider struct {
	clientID         string
	clientSecret     string
	webhookPublicKey string
	redirectURI      string
	authURL          string
	tokenURL         string
	userInfoURL      string
	scopes           []string
	httpClient       *http.Client
	paymentProvider  PaymentProviderClient
}

// NewHighLevelProvider creates a new HighLevel provider. webhookPublicKey is
// the PEM-encoded Ed25519 public key used to verify HighLevel webhook
// signatures (HIGHLEVEL_WEBHOOK_PUBLIC_KEY). It is public cryptographic
// material, not a private credential, and must not be confused with the OAuth
// client secret.
//
// paymentProvider is the Custom Payment Provider client used for outbound
// HighLevel provider registration/configuration calls. It may be nil if the
// provider does not support Custom Payment Provider operations.
func NewHighLevelProvider(clientID, clientSecret, redirectURI, webhookPublicKey string, paymentProvider PaymentProviderClient) *HighLevelProvider {
	return &HighLevelProvider{
		clientID:         clientID,
		clientSecret:     clientSecret,
		webhookPublicKey: webhookPublicKey,
		redirectURI:      redirectURI,
		authURL:          "https://api.highlevel.com/oauth/authorize",
		tokenURL:         "https://api.highlevel.com/oauth/token",
		userInfoURL:      "https://api.highlevel.com/v1/users/me",
		scopes:           []string{"read", "write"},
		// A single shared client is reused across all provider calls so HTTP
		// connections are pooled and reused rather than recreated per request.
		httpClient:      &http.Client{Timeout: 10 * time.Second},
		paymentProvider: paymentProvider,
	}
}

func (p *HighLevelProvider) ID() string {
	return "highlevel"
}

func (p *HighLevelProvider) Name() string {
	return "HighLevel"
}

func (p *HighLevelProvider) Capabilities() []Capability {
	caps := []Capability{
		CapabilityOAuth,
		CapabilityWebhooks,
		CapabilityTokenRefresh,
		CapabilityInstallation,
		CapabilityUninstallation,
	}
	if p.paymentProvider != nil {
		caps = append(caps, CapabilityPaymentProvider)
	}
	return caps
}

func (p *HighLevelProvider) HasCapability(capability Capability) bool {
	switch capability {
	case CapabilityOAuth, CapabilityWebhooks, CapabilityTokenRefresh, CapabilityInstallation, CapabilityUninstallation:
		return true
	case CapabilityPaymentProvider:
		return p.paymentProvider != nil
	default:
		return false
	}
}

func (p *HighLevelProvider) OAuthProvider() OAuthProvider {
	return p
}

func (p *HighLevelProvider) WebhookProvider() WebhookProvider {
	return NewHighLevelWebhookProvider(p.webhookPublicKey)
}

func (p *HighLevelProvider) PaymentProvider() PaymentProviderClient {
	return p.paymentProvider
}

func (p *HighLevelProvider) GenerateAuthorizationURL(ctx context.Context, state string, redirectURI string) (string, error) {
	u, err := url.Parse(p.authURL)
	if err != nil {
		return "", fmt.Errorf("invalid auth URL: %w", err)
	}

	q := u.Query()
	q.Set("client_id", p.clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", strings.Join(p.scopes, " "))
	q.Set("state", state)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (p *HighLevelProvider) ExchangeCode(ctx context.Context, code string, redirectURI string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", p.clientID)
	data.Set("client_secret", p.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", p.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &TokenResponse{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		TokenType:    tokenResp.TokenType,
		Scope:        tokenResp.Scope,
	}, nil
}

func (p *HighLevelProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", p.clientID)
	data.Set("client_secret", p.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", p.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed: %s", string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	return &TokenResponse{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		TokenType:    tokenResp.TokenType,
		Scope:        tokenResp.Scope,
	}, nil
}

func (p *HighLevelProvider) GetUserInfo(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.userInfoURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create user info request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("user info request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read user info response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("user info request failed: %s", string(body))
	}

	var userInfo struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return "", fmt.Errorf("failed to parse user info: %w", err)
	}

	return userInfo.ID, nil
}

func (p *HighLevelProvider) ValidateToken(ctx context.Context, accessToken string) (bool, error) {
	_, err := p.GetUserInfo(ctx, accessToken)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GenerateState creates a random state string for OAuth flow.
func GenerateState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
