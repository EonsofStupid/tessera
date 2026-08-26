-- Flows: login, MFA and recovery as declared configuration.
--
-- A flow is an ordered set of stages. The stages live in their own table
-- rather than a JSONB column on the flow because position is a real key —
-- "what is stage 2 of login-mfa" is a lookup, and reordering in a blueprint
-- diff shows as row changes a reviewer can read.

CREATE TYPE nomen_product.flow_designation AS ENUM (
    'authentication'
    , 'recovery'
);

CREATE TABLE nomen_product.flows (
    instance_id TEXT NOT NULL
    , slug TEXT NOT NULL CHECK (slug <> '')

    , title TEXT NOT NULL DEFAULT ''
    , designation nomen_product.flow_designation NOT NULL

    , created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    , updated_at TIMESTAMPTZ NOT NULL DEFAULT now()

    , PRIMARY KEY (instance_id, slug)
);

CREATE TABLE nomen_product.flow_stages (
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
        REFERENCES nomen_product.flows (instance_id, slug) ON DELETE CASCADE
);

-- One client's walk through a plan. Server-side on purpose: the client holds
-- an id and a token, never the plan, never the session token, until done.
CREATE TABLE nomen_product.flow_executions (
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
CREATE INDEX idx_flow_executions_expiry ON nomen_product.flow_executions (expires_at);

CREATE OR REPLACE TRIGGER trg_set_updated_at_flows
    BEFORE UPDATE ON nomen_product.flows
    FOR EACH ROW
    EXECUTE FUNCTION nomen_product.set_updated_at();
