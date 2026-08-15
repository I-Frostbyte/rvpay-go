package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/I-Frostbyte/rvpay-go/integrations/db/repo"
	"github.com/I-Frostbyte/rvpay-go/integrations/db/sqlc"
	"github.com/rs/zerolog"
)

const highLevelProvider = "highlevel"

type Service struct {
	repo   repo.IntegrationsRepo
	logger zerolog.Logger
}

func NewService(
	repo repo.IntegrationsRepo,
	logger zerolog.Logger,
) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

type highLevelWebhookPayload struct {
	Type      string          `json:"type"`
	EventType string          `json:"eventType"`
	Data      json.RawMessage `json:"data"`
}

func (s *Service) HandleWebhook(ctx context.Context, body []byte) error {
	if len(body) == 0 {
		return fmt.Errorf("webhook body is required")
	}

	var payload highLevelWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("invalid webhook payload: %w", err)
	}

	eventType := payload.EventType
	if strings.TrimSpace(eventType) == "" {
		eventType = payload.Type
	}
	if strings.TrimSpace(eventType) == "" {
		return fmt.Errorf("webhook event type is required")
	}

	queries := s.repo.Do()
	_, err := queries.CreateWebhookEvent(ctx, sqlc.CreateWebhookEventParams{
		Provider:  highLevelProvider,
		EventType: eventType,
		Payload:   body,
		Processed: false,
	})
	if err != nil {
		return fmt.Errorf("could not store webhook event: %w", err)
	}

	// Process asynchronously after the event is stored.
	go s.processEvent(eventType, body)

	return nil
}

func (s *Service) processEvent(eventType string, body []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.logger.Info().Str("event_type", eventType).Msg("processing webhook event asynchronously")

	// TODO: internal service communication will be added later.
	// Do not call deposits directly.
	_ = ctx
	_ = body
}

func (s *Service) ReadBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}
