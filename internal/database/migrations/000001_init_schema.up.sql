CREATE TABLE IF NOT EXISTS transaction_ledgers (
    id BIGSERIAL PRIMARY KEY,
    quantity BIGINT NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payment_gateway_communications (
    id BIGSERIAL PRIMARY KEY,
    transaction_ledger_id BIGINT NOT NULL,
    gateway_name VARCHAR(255) NOT NULL,
    operation VARCHAR(255) NOT NULL,
    request_json JSONB NOT NULL,
    request_timestamp TIMESTAMPTZ NOT NULL,
    response_json JSONB,
    response_status VARCHAR(255) NOT NULL,
    response_timestamp TIMESTAMPTZ NOT NULL,
    poll_attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_gateway_communications_gateway_name
    ON payment_gateway_communications (gateway_name);
CREATE INDEX IF NOT EXISTS idx_payment_gateway_communications_operation
    ON payment_gateway_communications (operation);
CREATE INDEX IF NOT EXISTS idx_payment_gateway_communications_response_status
    ON payment_gateway_communications (response_status);

CREATE TABLE IF NOT EXISTS gateway_priorities (
    id BIGSERIAL PRIMARY KEY,
    sort_order INT NOT NULL,
    gateway_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_gateway_priorities_sort_order UNIQUE (sort_order),
    CONSTRAINT uq_gateway_priorities_gateway_name UNIQUE (gateway_name)
);
