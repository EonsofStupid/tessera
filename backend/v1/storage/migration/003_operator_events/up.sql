CREATE TABLE tessera.tessera_operator_events (
    instance_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    event_id UUID NOT NULL,
    session_id UUID NOT NULL,
    sequence BIGINT NOT NULL CHECK (sequence > 0),
    occurred_at TIMESTAMPTZ NOT NULL,
    route_id TEXT NOT NULL,
    control_id TEXT NOT NULL DEFAULT '',
    event_type TEXT NOT NULL,
    action_id TEXT NOT NULL DEFAULT '',
    resource_revision TEXT NOT NULL DEFAULT '',
    correlation_id TEXT NOT NULL DEFAULT '',
    outcome TEXT NOT NULL DEFAULT '',
    attributes JSONB NOT NULL DEFAULT '{}',
    actor_id TEXT NOT NULL,
    agent_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (instance_id, tenant_id, event_id),
    UNIQUE (instance_id, tenant_id, session_id, sequence)
);

CREATE INDEX tessera_operator_events_tenant_time
    ON tessera.tessera_operator_events (instance_id, tenant_id, occurred_at DESC);

CREATE TABLE tessera.tessera_outbox (
    instance_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    event_id UUID NOT NULL,
    topic TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ,
    attempts INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (instance_id, tenant_id, event_id, topic)
);

CREATE INDEX tessera_outbox_pending
    ON tessera.tessera_outbox (created_at)
    WHERE published_at IS NULL;

ALTER TABLE tessera.tessera_operator_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE tessera.tessera_operator_events FORCE ROW LEVEL SECURITY;
ALTER TABLE tessera.tessera_outbox ENABLE ROW LEVEL SECURITY;
ALTER TABLE tessera.tessera_outbox FORCE ROW LEVEL SECURITY;

CREATE POLICY tessera_operator_events_tenant_policy ON tessera.tessera_operator_events
    USING (
        instance_id = current_setting('tessera.instance_id', true)
        AND tenant_id = current_setting('tessera.tenant_id', true)
    )
    WITH CHECK (
        instance_id = current_setting('tessera.instance_id', true)
        AND tenant_id = current_setting('tessera.tenant_id', true)
    );

CREATE POLICY tessera_outbox_tenant_policy ON tessera.tessera_outbox
    USING (
        instance_id = current_setting('tessera.instance_id', true)
        AND tenant_id = current_setting('tessera.tenant_id', true)
    )
    WITH CHECK (
        instance_id = current_setting('tessera.instance_id', true)
        AND tenant_id = current_setting('tessera.tenant_id', true)
    );
