CREATE TABLE payment_provider_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    integration_id UUID NOT NULL,

    provider_name TEXT NOT NULL,

    provider_description TEXT,

    provider_image_url TEXT,

    location_id TEXT NOT NULL,

    query_url TEXT NOT NULL,

    payments_url TEXT NOT NULL,

    supports_subscription_schedule BOOLEAN NOT NULL DEFAULT FALSE,

    provider_api_key TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_payment_provider_config_integration
        FOREIGN KEY (integration_id)
        REFERENCES integrations(id)
        ON DELETE CASCADE
        ON UPDATE RESTRICT,

    CONSTRAINT uq_payment_provider_config_integration UNIQUE (integration_id)
);

CREATE INDEX idx_payment_provider_configs_location_id ON payment_provider_configs (location_id);