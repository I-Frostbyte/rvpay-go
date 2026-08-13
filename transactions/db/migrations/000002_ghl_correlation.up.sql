ALTER TABLE deposits
    ADD COLUMN ghl_transaction_id TEXT,
    ADD COLUMN ghl_charge_id TEXT;

CREATE INDEX idx_deposits_ghl_transaction_id ON deposits (ghl_transaction_id);

CREATE INDEX idx_deposits_ghl_charge_id ON deposits (ghl_charge_id);