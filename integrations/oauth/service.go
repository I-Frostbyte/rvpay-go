package oauth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/I-Frostbyte/rvpay-go/integrations/db/repo"
	"github.com/I-Frostbyte/rvpay-go/integrations/db/sqlc"
	"github.com/rs/zerolog"
)

const highLevelTokenURL = "https://services.leadconnectorhq.com/oauth/token"

type Service struct {
	repo            repo.IntegrationsRepo
	logger          zerolog.Logger
	clientID        string
	clientSecret    string
	redirectURL     string
	encryptionKey   []byte
	httpClient      *http.Client
}

func NewService(
	repo repo.IntegrationsRepo,
	logger zerolog.Logger,
	clientID string,
	clientSecret string,
	redirectURL string,
	encryptionKey string,
) *Service {
	return &Service{
		repo:          repo,
		logger:        logger,
		clientID:      clientID,
		clientSecret:  clientSecret,
		redirectURL:   redirectURL,
		encryptionKey: []byte(encryptionKey),
		httpClient:    &http.Client{Timeout: 10 * time.Second},
	}
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	UserType     string `json:"userType"`
	CompanyID    string `json:"companyId"`
	LocationID   string `json:"locationId"`
	UserID       string `json:"userId"`
	PlanID       string `json:"planId"`
}

func (s *Service) HandleCallback(ctx context.Context, code string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("authorization code is required")
	}

	token, err := s.exchangeCode(ctx, code)
	if err != nil {
		return fmt.Errorf("could not exchange code with HighLevel: %w", err)
	}

	encryptedAccessToken, err := encryptToken(token.AccessToken, s.encryptionKey)
	if err != nil {
		return fmt.Errorf("could not encrypt access token: %w", err)
	}

	encryptedRefreshToken, err := encryptToken(token.RefreshToken, s.encryptionKey)
	if err != nil {
		return fmt.Errorf("could not encrypt refresh token: %w", err)
	}

	queries := s.repo.Do()
	_, err = queries.CreateIntegration(ctx, sqlc.CreateIntegrationParams{
		Provider:       "highlevel",
		LocationID:     token.LocationID,
		AccessToken:    encryptedAccessToken,
		RefreshToken:   encryptedRefreshToken,
		TokenExpiresAt: time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
	})
	if err != nil {
		return fmt.Errorf("could not store integration: %w", err)
	}

	return nil
}

func (s *Service) exchangeCode(ctx context.Context, code string) (*tokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", s.clientID)
	form.Set("client_secret", s.clientSecret)
	form.Set("redirect_uri", s.redirectURL)
	form.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, highLevelTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("highlevel returned status %d: %s", resp.StatusCode, string(body))
	}

	var token tokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

func encryptToken(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}