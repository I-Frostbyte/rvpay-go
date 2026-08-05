CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE payment_provider AS ENUM (
    'MTN_MOMO_CMR',
    'ORANGE_CMR'
);

CREATE TYPE payer_type AS ENUM (
    'MMO'
);

CREATE TABLE clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    client_name TEXT NOT NULL,

    email TEXT NOT NULL UNIQUE,

    phone_number TEXT NOT NULL UNIQUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE deposits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    amount NUMERIC(18,2) NOT NULL,

    currency VARCHAR(3) NOT NULL
        CHECK (currency ~ '^[A-Z]{3}$'),

    payer_type payer_type NOT NULL,

    payer_phone_number TEXT NOT NULL,

    payer_provider payment_provider NOT NULL,

    client_id UUID NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_deposit_client
        FOREIGN KEY (client_id)
        REFERENCES clients(id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);
