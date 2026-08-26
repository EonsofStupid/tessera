CREATE TABLE nomen.session_user_agents (
    instance_id TEXT NOT NULL
    , fingerprint_id TEXT NOT NULL CHECK (fingerprint_id <> '')
    , ip INET
    , description TEXT
    , headers JSONB

    , PRIMARY KEY (instance_id, fingerprint_id)
);

CREATE TABLE nomen.sessions (
    instance_id TEXT NOT NULL
    , id TEXT NOT NULL CHECK (id <> '')
    , token_id TEXT
    , user_agent_id TEXT
    , lifetime INTERVAL
    , expiration TIMESTAMPTZ
    , user_id TEXT -- this column is used for referential integrity
    , creator_id TEXT
    , created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
    , updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL

    , PRIMARY KEY (instance_id, id)
    , FOREIGN KEY (instance_id) REFERENCES nomen.instances(id)
    , FOREIGN KEY (instance_id, user_agent_id) REFERENCES nomen.session_user_agents(instance_id, fingerprint_id) ON DELETE SET NULL (user_agent_id)
);

CREATE OR REPLACE FUNCTION nomen.update_expiration()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.expiration := NEW.updated_at + NEW.lifetime;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_expiration_on_update
    BEFORE UPDATE OF lifetime ON nomen.sessions
    FOR EACH ROW
EXECUTE FUNCTION nomen.update_expiration();

CREATE TRIGGER set_expiration_on_insert
    BEFORE INSERT ON nomen.sessions
    FOR EACH ROW
    WHEN (NEW.lifetime <> '0'::interval)
EXECUTE FUNCTION nomen.update_expiration();

CREATE TYPE nomen.session_factor_type AS ENUM (
    'user',
    'password',
    'passkey', -- is also a challenge
    'identity_provider_intent',
    'totp',
    'otp_sms', -- is also a challenge
    'otp_email', -- is also a challenge
    'recovery_code'
);

CREATE TABLE nomen.session_factors (
    instance_id TEXT NOT NULL
    , session_id TEXT NOT NULL
    , type nomen.session_factor_type NOT NULL
    , last_challenged_at TIMESTAMPTZ -- this is only set if the type is a challenge
    , challenged_payload JSONB
    , last_verified_at TIMESTAMPTZ
    , verified_payload JSONB

    , PRIMARY KEY (instance_id, session_id, type)
    , FOREIGN KEY (instance_id, session_id) REFERENCES nomen.sessions(instance_id, id) ON DELETE CASCADE
);

CREATE TABLE nomen.session_metadata (
    instance_id TEXT NOT NULL
    , session_id TEXT NOT NULL
    , key TEXT NOT NULL CHECK (key <> '')
    , value BYTEA NOT NULL

    , created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    , updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()

    , PRIMARY KEY (instance_id, session_id, key)

    , CONSTRAINT fk_session_metadata_session FOREIGN KEY (instance_id, session_id) REFERENCES nomen.sessions (instance_id, id) ON DELETE CASCADE
);

CREATE INDEX idx_session_metadata_key ON nomen.session_metadata (key);
CREATE INDEX idx_session_metadata_value ON nomen.session_metadata (sha256(value));

-- TODO(adlerhurst): these indexes can currently not be used by Postgres, because of the type conversion
-- the value can be a json but doesn't have to be.
-- CREATE INDEX idx_session_metadata_value_number ON nomen.session_metadata ((value::NUMERIC)) WHERE jsonb_typeof(value) = 'number';
-- CREATE INDEX idx_session_metadata_value_string ON nomen.session_metadata ((value#>>'{}')) WHERE jsonb_typeof(value) = 'string';
-- CREATE INDEX idx_session_metadata_value_boolean ON nomen.session_metadata ((value::BOOLEAN)) WHERE jsonb_typeof(value) = 'boolean';

CREATE TRIGGER trg_set_updated_at_session_metadata
    BEFORE INSERT OR UPDATE ON nomen.session_metadata
    FOR EACH ROW
    WHEN (NEW.updated_at IS NULL)
EXECUTE FUNCTION nomen.set_updated_at();