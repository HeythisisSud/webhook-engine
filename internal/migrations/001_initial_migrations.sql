-- +goose Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE webhooks (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id   TEXT NOT NULL,
    target_url  TEXT NOT NULL,
    secret      TEXT NOT NULL,
    event_types TEXT[] NOT NULL DEFAULT '{}',
    enabled     BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhooks_client_id ON webhooks(client_id);
CREATE INDEX idx_webhooks_event_types ON webhooks USING GIN(event_types);





CREATE TABLE events (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id   TEXT NOT NULL,
    event_type  TEXT NOT NULL,
    payload     JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_client_id ON events(client_id);
CREATE INDEX idx_events_event_type ON events(event_type);

CREATE TABLE outbox (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id    UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    webhook_id  UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    status      TEXT NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outbox_status ON outbox(status);
CREATE INDEX idx_outbox_event_id ON outbox(event_id);

CREATE TABLE delivery_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    outbox_id       UUID NOT NULL REFERENCES outbox(id) ON DELETE CASCADE,
    attempt_number  INT NOT NULL DEFAULT 1,
    status_code     INT,
    response_body   TEXT,
    error_message   TEXT,
    success         BOOLEAN NOT NULL DEFAULT false,
    delivered_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_delivery_logs_outbox_id ON delivery_logs(outbox_id);

-- +goose Down

DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS delivery_logs;
DROP TABLE IF EXISTS outbox;
DROP TABLE IF EXISTS events;
