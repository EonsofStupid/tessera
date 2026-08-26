CREATE TABLE nomen.organization_metadata (
    instance_id TEXT NOT NULL
    , organization_id TEXT NOT NULL
    , key TEXT NOT NULL CHECK (key <> '')
    , value BYTEA NOT NULL

    , created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    , updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    
    , PRIMARY KEY (instance_id, organization_id, key)
    
    , CONSTRAINT fk_organization_metadata_org FOREIGN KEY (instance_id, organization_id) REFERENCES nomen.organizations (instance_id, id) ON DELETE CASCADE
);

CREATE INDEX idx_organization_metadata_key ON nomen.organization_metadata (key);
CREATE INDEX idx_organization_metadata_value ON nomen.organization_metadata (sha256(value));

-- TODO(adlerhurst): these indexes can currently not be used by Postgres, because of the type conversion
-- the value can be a json but doesn't have to be.
-- CREATE INDEX idx_organization_metadata_value_number ON nomen.organization_metadata ((value::NUMERIC)) WHERE jsonb_typeof(value) = 'number';
-- CREATE INDEX idx_organization_metadata_value_string ON nomen.organization_metadata ((value#>>'{}')) WHERE jsonb_typeof(value) = 'string';
-- CREATE INDEX idx_organization_metadata_value_boolean ON nomen.organization_metadata ((value::BOOLEAN)) WHERE jsonb_typeof(value) = 'boolean';

CREATE TRIGGER trg_set_updated_at_organization_metadata
  BEFORE INSERT OR UPDATE ON nomen.organization_metadata
  FOR EACH ROW
  WHEN (NEW.updated_at IS NULL)
  EXECUTE FUNCTION nomen.set_updated_at();