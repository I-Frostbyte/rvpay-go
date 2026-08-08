DROP INDEX IF EXISTS idx_platforms_enabled;

DROP INDEX IF EXISTS idx_integrations_platform_id;

DROP INDEX IF EXISTS idx_integrations_external_account_id;

DROP INDEX IF EXISTS idx_integrations_status;

DROP INDEX IF EXISTS idx_webhook_subscriptions_status;

ALTER TABLE IF EXISTS platform_metadata DROP CONSTRAINT IF EXISTS fk_platform_metadata_platform;

ALTER TABLE IF EXISTS platform_metadata DROP CONSTRAINT IF EXISTS uq_platform_metadata_key;

ALTER TABLE IF EXISTS webhook_subscriptions DROP CONSTRAINT IF EXISTS fk_webhook_subscription_integration;

ALTER TABLE IF EXISTS webhook_subscriptions DROP CONSTRAINT IF EXISTS uq_webhook_subscription_integration_endpoint;

ALTER TABLE IF EXISTS oauth_tokens DROP CONSTRAINT IF EXISTS fk_oauth_token_integration;

ALTER TABLE IF EXISTS oauth_tokens DROP CONSTRAINT IF EXISTS uq_oauth_token_integration;

ALTER TABLE IF EXISTS integrations DROP CONSTRAINT IF EXISTS fk_integration_client;

ALTER TABLE IF EXISTS integrations DROP CONSTRAINT IF EXISTS fk_integration_platform;

ALTER TABLE IF EXISTS integrations DROP CONSTRAINT IF EXISTS uq_integration_client_platform;

DROP TABLE IF EXISTS platform_metadata;

DROP TABLE IF EXISTS webhook_subscriptions;

DROP TABLE IF EXISTS oauth_tokens;

DROP TABLE IF EXISTS integrations;

DROP TABLE IF EXISTS platforms;

DROP TABLE IF EXISTS clients;

DROP TYPE IF EXISTS webhook_subscription_status;

DROP TYPE IF EXISTS integration_status;

DROP TYPE IF EXISTS client_status;