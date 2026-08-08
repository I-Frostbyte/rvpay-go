CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE client_status AS ENUM (
    'REGISTERED',
    'ACTIVE',
    'SUSPENDED',
    'CLOSED'
);

CREATE TYPE integration_status AS ENUM (
    'CREATED',
    'OAUTH_PENDING',
    'ACTIVE',
    'REVOKED'
);

CREATE TYPE webhook_subscription_status AS ENUM (
    'ACTIVE',
    'DISABLED'
);

CREATE TABLE clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    client_name TEXT NOT NULL,

    status client_status NOT NULL DEFAULT 'REGISTERED',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE platforms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name TEXT NOT NULL,

    display_name TEXT NOT NULL,

    slug TEXT NOT NULL UNIQUE,

    enabled BOOLEAN NOT NULL DEFAULT TRUE,

    oauth_capable BOOLEAN NOT NULL DEFAULT FALSE,

    webhook_capable BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_platforms_enabled ON platforms (enabled);

CREATE TABLE integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    client_id UUID NOT NULL,

    platform_id UUID NOT NULL,

    external_account_id TEXT,

    status integration_status NOT NULL DEFAULT 'CREATED',

    installed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    last_sync_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_integration_client
        FOREIGN KEY (client_id)
        REFERENCES clients(id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT,

    CONSTRAINT fk_integration_platform
        FOREIGN KEY (platform_id)
        REFERENCES platforms(id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT,

    CONSTRAINT uq_integration_client_platform UNIQUE (client_id, platform_id)
);

CREATE INDEX idx_integrations_platform_id ON integrations (platform_id);

CREATE INDEX idx_integrations_external_account_id ON integrations (external_account_id);

CREATE INDEX idx_integrations_status ON integrations (status);

CREATE TABLE oauth_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    integration_id UUID NOT NULL,

    access_token TEXT NOT NULL,

    refresh_token TEXT NOT NULL,

    expires_at TIMESTAMPTZ NOT NULL,

    scope TEXT,

    token_type TEXT NOT NULL DEFAULT 'bearer',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_oauth_token_integration
        FOREIGN KEY (integration_id)
        REFERENCES integrations(id)
        ON DELETE CASCADE
        ON UPDATE RESTRICT,

    CONSTRAINT uq_oauth_token_integration UNIQUE (integration_id)
);

CREATE TABLE webhook_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    integration_id UUID NOT NULL,

    endpoint TEXT NOT NULL,

    secret TEXT NOT NULL,

    status webhook_subscription_status NOT NULL DEFAULT 'ACTIVE',

    last_delivery TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_webhook_subscription_integration
        FOREIGN KEY (integration_id)
        REFERENCES integrations(id)
        ON DELETE CASCADE
        ON UPDATE RESTRICT,

    CONSTRAINT uq_webhook_subscription_integration_endpoint UNIQUE (integration_id, endpoint)
);

CREATE INDEX idx_webhook_subscriptions_status ON webhook_subscriptions (status);

CREATE TABLE platform_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    platform_id UUID NOT NULL,

    metadata_key TEXT NOT NULL,

    metadata_value TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_platform_metadata_platform
        FOREIGN KEY (platform_id)
        REFERENCES platforms(id)
        ON DELETE CASCADE
        ON UPDATE RESTRICT,

    CONSTRAINT uq_platform_metadata_key UNIQUE (platform_id, metadata_key)
);