CREATE TABLE oauth_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    state TEXT NOT NULL UNIQUE,

    client_id UUID NOT NULL,

    platform_id UUID NOT NULL,

    expires_at TIMESTAMPTZ NOT NULL,

    consumed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_oauth_state_client
        FOREIGN KEY (client_id)
        REFERENCES clients(id)
        ON DELETE CASCADE
        ON UPDATE RESTRICT,

    CONSTRAINT fk_oauth_state_platform
        FOREIGN KEY (platform_id)
        REFERENCES platforms(id)
        ON DELETE CASCADE
        ON UPDATE RESTRICT
);

CREATE INDEX idx_oauth_states_state ON oauth_states (state);

CREATE INDEX idx_oauth_states_expires_at ON oauth_states (expires_at);