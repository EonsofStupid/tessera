-- Flows: login, MFA and recovery as declared configuration.
--
-- A flow is an ordered set of stages. The stages live in their own table
-- rather than a JSONB column on the flow because position is a real key —
-- "what is stage 2 of login-mfa" is a lookup, and reordering in a blueprint
-- diff shows as row changes a reviewer can read.

CREATE TYPE tessera.flow_designation AS ENUM (
    'authentication'
    , 'recovery'
);

CREATE TABLE tessera.flows (
    instance_id TEXT NOT NULL
    , slug TEXT NOT NULL CHECK (slug <> '')

    , title TEXT NOT NULL DEFAULT ''
    , designation tessera.flow_designation NOT NULL

    , created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    , updated_at TIMESTAMPTZ NOT NULL DEFAULT now()

    , PRIMARY KEY (instance_id, slug)
);

CREATE TABLE tessera.flow_stages (
    instance_id TEXT NOT NULL
    , flow_slug TEXT NOT NULL
    , position SMALLINT NOT NULL CHECK (position >= 0)

    , kind TEXT NOT NULL CHECK (kind <> '')
    -- Stage-specific configuration, decoded strictly by the stage
    -- implementation. JSONB rather than columns because kinds differ in
    -- shape and the set of kinds will grow.
    , config JSONB NOT NULL DEFAULT '{}'

    , PRIMARY KEY (instance_id, flow_slug, position)
    , CONSTRAINT fk_stage_flow FOREIGN KEY (instance_id, flow_slug)
        REFERENCES tessera.flows (instance_id, slug) ON DELETE CASCADE
);

-- One client's walk through a plan. Server-side on purpose: the client holds
-- an id and a token, never the plan, never the session token, until done.
CREATE TABLE tessera.flow_executions (
    instance_id TEXT NOT NULL
    , id TEXT NOT NULL

    , token TEXT NOT NULL
    , plan JSONB NOT NULL
    , position SMALLINT NOT NULL DEFAULT 0

    , user_id TEXT NOT NULL DEFAULT ''
    , session_id TEXT NOT NULL DEFAULT ''
    , session_token TEXT NOT NULL DEFAULT ''
    , resource_owner TEXT NOT NULL DEFAULT ''

    , created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    , expires_at TIMESTAMPTZ NOT NULL

    , PRIMARY KEY (instance_id, id)
);

-- Expiry sweeps scan by time, not by tenant.
CREATE INDEX idx_flow_executions_expiry ON tessera.flow_executions (expires_at);

CREATE OR REPLACE TRIGGER trg_set_updated_at_flows
    BEFORE UPDATE ON tessera.flows
    FOR EACH ROW
    EXECUTE FUNCTION tessera.set_updated_at();
