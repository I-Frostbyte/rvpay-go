DROP INDEX IF EXISTS idx_deposits_ghl_transaction_id;

DROP INDEX IF EXISTS idx_deposits_ghl_charge_id;

ALTER TABLE deposits
    DROP COLUMN IF EXISTS ghl_transaction_id,
    DROP COLUMN IF EXISTS ghl_charge_id;