DROP INDEX IF EXISTS idx_payment_provider_configs_location_id;

ALTER TABLE IF EXISTS payment_provider_configs DROP CONSTRAINT IF EXISTS fk_payment_provider_config_integration;

ALTER TABLE IF EXISTS payment_provider_configs DROP CONSTRAINT IF EXISTS uq_payment_provider_config_integration;

DROP TABLE IF EXISTS payment_provider_configs;