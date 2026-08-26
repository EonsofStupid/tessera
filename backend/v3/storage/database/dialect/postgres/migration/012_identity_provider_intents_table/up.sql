CREATE TYPE nomen.idp_intent_state AS ENUM (
    'started',
    'succeeded',
    'failed',
    'consumed'
    );

CREATE TABLE nomen.identity_provider_intents
(
    instance_id             TEXT                        NOT NULL,
    id                      TEXT                        NOT NULL CHECK ( id <> '' ),
    state                   nomen.idp_intent_state    NOT NULL DEFAULT 'started',
    success_url             TEXT,
    failure_url             TEXT,
    created_at              TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    idp_id                  TEXT                        NOT NULL,
    idp_arguments           JSONB,
    idp_user                BYTEA, 
    idp_user_id             TEXT,
    idp_username            TEXT,
    user_id                 TEXT,
    idp_access_token        JSONB,
    idp_id_token            TEXT,
    idp_entry_attributes    JSONB,
    request_id              TEXT,
    assertion               JSONB,
    succeeded_at            TIMESTAMPTZ,
    fail_reason             TEXT,
    failed_at               TIMESTAMPTZ,
    expires_at              TIMESTAMPTZ,

    PRIMARY KEY (instance_id, id),
    FOREIGN KEY (instance_id, idp_id) REFERENCES nomen.identity_providers (instance_id, id) ON DELETE CASCADE
);

CREATE TRIGGER trigger_set_updated_at
    BEFORE UPDATE
    ON nomen.identity_provider_intents
    FOR EACH ROW
    WHEN (NEW.updated_at IS NULL)
EXECUTE FUNCTION nomen.set_updated_at();