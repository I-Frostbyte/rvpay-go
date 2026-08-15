package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HighLevelPaymentProviderClient implements PaymentProviderClient for
// HighLevel's Custom Payment Provider API. It makes authenticated outbound
// calls to HighLevel using the installed location's OAuth access token.
//
// The access token is used only in the Authorization header and is never
// logged or included in returned errors.
type HighLevelPaymentProviderClient struct {
	// baseURL is the HighLevel API base URL (e.g. https://services.leadconnectorhq.com).
	baseURL string
	// httpClient is the shared HTTP client used for outbound calls.
	httpClient *http.Client
}

// NewHighLevelPaymentProviderClient creates a new HighLevel payment provider
// client. baseURL is the HighLevel API base URL; it must not include a
// trailing slash. httpClient may be nil, in which case a default client with
// a 10-second timeout is used.
func NewHighLevelPaymentProviderClient(baseURL string, httpClient *http.Client) *HighLevelPaymentProviderClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &HighLevelPaymentProviderClient{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: httpClient,
	}
}

// CreateProviderAssociation creates the association between the Marketplace
// app and the HighLevel location.
//
// POST /payments/custom-provider/provider
func (c *HighLevelPaymentProviderClient) CreateProviderAssociation(ctx context.Context, accessToken, locationID string) error {
	if strings.TrimSpace(accessToken) == "" {
		return ErrMissingAccessToken
	}
	if strings.TrimSpace(locationID) == "" {
		return ErrMissingLocationID
	}

	body := map[string]string{
		"locationId": locationID,
	}

	var respBody map[string]interface{}
	if err := c.doJSON(ctx, http.MethodPost, "/payments/custom-provider/provider", accessToken, body, &respBody); err != nil {
		return err
	}

	return nil
}

// CreateProviderConfig creates the provider configuration for a location.
//
// POST /payments/custom-provider/connect
func (c *HighLevelPaymentProviderClient) CreateProviderConfig(ctx context.Context, accessToken string, config ProviderConfig) error {
	if strings.TrimSpace(accessToken) == "" {
		return ErrMissingAccessToken
	}
	if strings.TrimSpace(config.LocationID) == "" {
		return ErrMissingLocationID
	}

	body := map[string]interface{}{
		"name":                         config.Name,
		"description":                  config.Description,
		"imageUrl":                     config.ImageURL,
		"locationId":                   config.LocationID,
		"queryUrl":                     config.QueryURL,
		"paymentsUrl":                  config.PaymentsURL,
		"supportsSubscriptionSchedule": config.SupportsSubscriptionSchedule,
	}

	var respBody map[string]interface{}
	if err := c.doJSON(ctx, http.MethodPost, "/payments/custom-provider/connect", accessToken, body, &respBody); err != nil {
		return err
	}

	return nil
}

// FetchProviderConfig fetches the provider configuration for a location.
//
// GET /payments/custom-provider/connect
func (c *HighLevelPaymentProviderClient) FetchProviderConfig(ctx context.Context, accessToken, locationID string) (*ProviderConfig, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, ErrMissingAccessToken
	}
	if strings.TrimSpace(locationID) == "" {
		return nil, ErrMissingLocationID
	}

	path := "/payments/custom-provider/connect"
	q := url.Values{}
	q.Set("locationId", locationID)
	fullPath := path + "?" + q.Encode()

	var respBody struct {
		Name                         string `json:"name"`
		Description                  string `json:"description"`
		ImageURL                     string `json:"imageUrl"`
		LocationID                   string `json:"locationId"`
		QueryURL                     string `json:"queryUrl"`
		PaymentsURL                  string `json:"paymentsUrl"`
		SupportsSubscriptionSchedule bool   `json:"supportsSubscriptionSchedule"`
	}

	if err := c.doJSON(ctx, http.MethodGet, fullPath, accessToken, nil, &respBody); err != nil {
		return nil, err
	}

	return &ProviderConfig{
		Name:                         respBody.Name,
		Description:                  respBody.Description,
		ImageURL:                     respBody.ImageURL,
		LocationID:                   respBody.LocationID,
		QueryURL:                     respBody.QueryURL,
		PaymentsURL:                  respBody.PaymentsURL,
		SupportsSubscriptionSchedule: respBody.SupportsSubscriptionSchedule,
	}, nil
}

// DisconnectProvider disconnects the provider configuration for a location.
//
// DELETE /payments/custom-provider/connect
func (c *HighLevelPaymentProviderClient) DisconnectProvider(ctx context.Context, accessToken, locationID string) error {
	if strings.TrimSpace(accessToken) == "" {
		return ErrMissingAccessToken
	}
	if strings.TrimSpace(locationID) == "" {
		return ErrMissingLocationID
	}

	path := "/payments/custom-provider/connect"
	q := url.Values{}
	q.Set("locationId", locationID)
	fullPath := path + "?" + q.Encode()

	var respBody map[string]interface{}
	if err := c.doJSON(ctx, http.MethodDelete, fullPath, accessToken, nil, &respBody); err != nil {
		return err
	}

	return nil
}

// doJSON performs an authenticated JSON request to the HighLevel API. It
// handles 2xx, 400, 401, and 422 responses and returns typed/domain errors.
// The access token is never logged or included in returned errors.
func (c *HighLevelPaymentProviderClient) doJSON(ctx context.Context, method, path, accessToken string, body, out interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("request cancelled: %w", ctx.Err())
		}
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		if out != nil && len(respData) > 0 {
			if err := json.Unmarshal(respData, out); err != nil {
				return fmt.Errorf("parse response body: %w", err)
			}
		}
		return nil
	case resp.StatusCode == http.StatusBadRequest:
		return fmt.Errorf("%w: %s", ErrBadRequest, sanitizeErrorBody(respData))
	case resp.StatusCode == http.StatusUnauthorized:
		return fmt.Errorf("%w: %s", ErrUnauthorized, sanitizeErrorBody(respData))
	case resp.StatusCode == http.StatusUnprocessableEntity:
		return fmt.Errorf("%w: %s", ErrUnprocessableEntity, sanitizeErrorBody(respData))
	default:
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, sanitizeErrorBody(respData))
	}
}

// credentialFields are JSON field names whose values must be redacted from
// error messages to prevent credential leakage.
var credentialFields = []string{
	"access_token",
	"refresh_token",
	"client_secret",
	"apiKey",
	"api_key",
}

// sanitizeErrorBody truncates and sanitizes an error response body so that
// sensitive data (including access tokens, refresh tokens, and API keys) is
// never leaked into errors. It attempts to parse the body as JSON and redact
// known credential fields. If parsing fails, it falls back to length-limited
// plain-text truncation.
func sanitizeErrorBody(body []byte) string {
	if len(body) == 0 {
		return "empty response body"
	}

	// Attempt to redact credential fields from JSON bodies.
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err == nil {
		for _, field := range credentialFields {
			if _, ok := raw[field]; ok {
				raw[field] = "[REDACTED]"
			}
		}
		sanitized, err := json.Marshal(raw)
		if err == nil {
			body = sanitized
		}
	}

	// Truncate to a reasonable length to avoid leaking large payloads.
	if len(body) > 512 {
		body = body[:512]
	}
	return string(body)
}

// Sentinel errors returned by the HighLevel payment provider client.
var (
	// ErrMissingAccessToken is returned when no access token is provided.
	ErrMissingAccessToken = errors.New("access token is required")
	// ErrMissingLocationID is returned when no location ID is provided.
	ErrMissingLocationID = errors.New("location ID is required")
	// ErrBadRequest is returned when HighLevel responds with 400.
	ErrBadRequest = errors.New("highlevel: bad request")
	// ErrUnauthorized is returned when HighLevel responds with 401.
	ErrUnauthorized = errors.New("highlevel: unauthorized")
	// ErrUnprocessableEntity is returned when HighLevel responds with 422.
	ErrUnprocessableEntity = errors.New("highlevel: unprocessable entity")
)
