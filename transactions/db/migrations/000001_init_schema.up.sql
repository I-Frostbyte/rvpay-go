CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE merchant_status AS ENUM (
    'ONBOARDED',
    'ACTIVE',
    'SUSPENDED',
    'RETIRED'
);

CREATE TYPE customer_status AS ENUM (
    'CREATED',
    'ACTIVE'
);

CREATE TYPE deposit_status AS ENUM (
    'INITIATED',
    'PROCESSING',
    'COMPLETED',
    'FAILED'
);

CREATE TYPE payout_status AS ENUM (
    'REQUESTED',
    'PROCESSING',
    'COMPLETED',
    'FAILED'
);

CREATE TYPE payment_provider AS ENUM (
    'MTN_MOMO',
    'ORANGE_MOMO'
);

CREATE TYPE payment_type AS ENUM (
    'MMO',
    'CREDIT_CARD'
);

CREATE TABLE merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name TEXT NOT NULL,

    slug TEXT NOT NULL UNIQUE,

    status merchant_status NOT NULL DEFAULT 'ONBOARDED',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    client_id UUID NOT NULL,

    merchant_id UUID NOT NULL,

    phone_number TEXT NOT NULL,

    status customer_status NOT NULL DEFAULT 'CREATED',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_customer_merchant
        FOREIGN KEY (merchant_id)
        REFERENCES merchants(id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT,

    CONSTRAINT uq_customer_client_merchant_phone UNIQUE (client_id, merchant_id, phone_number)
);

CREATE INDEX idx_customers_client_id ON customers (client_id);

CREATE INDEX idx_customers_merchant_id ON customers (merchant_id);

CREATE TABLE deposits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    client_id UUID NOT NULL,

    customer_id UUID NOT NULL,

    merchant_id UUID NOT NULL,

    amount NUMERIC(18,2) NOT NULL,

    currency VARCHAR(3) NOT NULL
        CHECK (currency ~ '^[A-Z]{3}$'),

    payment_type payment_type NOT NULL,

    payer_phone_number TEXT NOT NULL,

    provider payment_provider NOT NULL,

    status deposit_status NOT NULL DEFAULT 'INITIATED',

    external_reference TEXT,

    idempotency_key UUID NOT NULL UNIQUE,

    initiated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    completed_at TIMESTAMPTZ,

    failed_at TIMESTAMPTZ,

    failure_reason TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_deposit_customer
        FOREIGN KEY (customer_id)
        REFERENCES customers(id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT,

    CONSTRAINT fk_deposit_merchant
        FOREIGN KEY (merchant_id)
        REFERENCES merchants(id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);

CREATE INDEX idx_deposits_client_id ON deposits (client_id);

CREATE INDEX idx_deposits_customer_id ON deposits (customer_id);

CREATE INDEX idx_deposits_merchant_id ON deposits (merchant_id);

CREATE INDEX idx_deposits_status ON deposits (status);

CREATE INDEX idx_deposits_external_reference ON deposits (external_reference);

CREATE INDEX idx_deposits_created_at ON deposits (created_at);

CREATE TABLE payouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    client_id UUID NOT NULL,

    merchant_id UUID NOT NULL,

    amount NUMERIC(18,2) NOT NULL,

    currency VARCHAR(3) NOT NULL
        CHECK (currency ~ '^[A-Z]{3}$'),

    provider payment_provider NOT NULL,

    destination_reference TEXT,

    status payout_status NOT NULL DEFAULT 'REQUESTED',

    external_reference TEXT,

    idempotency_key UUID NOT NULL UNIQUE,

    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    completed_at TIMESTAMPTZ,

    failed_at TIMESTAMPTZ,

    failure_reason TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_payout_merchant
        FOREIGN KEY (merchant_id)
        REFERENCES merchants(id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);

CREATE INDEX idx_payouts_client_id ON payouts (client_id);

CREATE INDEX idx_payouts_merchant_id ON payouts (merchant_id);

CREATE INDEX idx_payouts_status ON payouts (status);

CREATE INDEX idx_payouts_external_reference ON payouts (external_reference);

CREATE INDEX idx_payouts_created_at ON payouts (created_at);