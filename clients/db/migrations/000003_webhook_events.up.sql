CREATE TABLE webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    integration_id UUID NOT NULL,

    provider_event_id TEXT NOT NULL,

    event_type TEXT NOT NULL,

    payload JSONB NOT NULL,

    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_webhook_event_integration
        FOREIGN KEY (integration_id)
        REFERENCES integrations(id)
        ON DELETE CASCADE
        ON UPDATE RESTRICT,

    CONSTRAINT uq_webhook_event_integration_provider UNIQUE (integration_id, provider_event_id)
);

CREATE INDEX idx_webhook_events_integration_id ON webhook_events (integration_id);

CREATE INDEX idx_webhook_events_provider_event_id ON webhook_events (provider_event_id);