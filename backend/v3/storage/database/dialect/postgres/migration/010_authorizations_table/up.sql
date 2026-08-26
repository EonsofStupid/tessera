CREATE TYPE nomen.authorization_state AS ENUM (
    'active',
    'inactive'
    );

CREATE TABLE nomen.authorizations
(
    instance_id TEXT                        NOT NULL,
    id          TEXT                        NOT NULL CHECK ( id <> '' ),
    state       nomen.authorization_state NOT NULL DEFAULT 'active',
    project_id  TEXT                        NOT NULL,
    grant_id    TEXT,
    user_id     TEXT                        NOT NULL,
    created_at  TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    PRIMARY KEY (instance_id, id),
    FOREIGN KEY (instance_id, project_id) REFERENCES nomen.projects (instance_id, id) ON DELETE CASCADE,
    FOREIGN KEY (instance_id, grant_id) REFERENCES nomen.project_grants (instance_id, id) ON DELETE CASCADE
);

CREATE TABLE nomen.authorization_roles
(
    instance_id      TEXT NOT NULL,
    authorization_id TEXT NOT NULL,
    role_key         TEXT NOT NULL CHECK ( role_key <> '' ),
    project_id       TEXT NOT NULL,
    grant_id         TEXT,
    PRIMARY KEY (instance_id, authorization_id, role_key),
    FOREIGN KEY (instance_id, authorization_id) REFERENCES nomen.authorizations (instance_id, id) ON DELETE CASCADE,
    FOREIGN KEY (instance_id, project_id, role_key) REFERENCES nomen.project_roles (instance_id, project_id, key) ON DELETE CASCADE,
    FOREIGN KEY (instance_id, grant_id, role_key) REFERENCES nomen.project_grant_roles (instance_id, grant_id, key) ON DELETE CASCADE
);

CREATE TRIGGER trigger_set_updated_at
    BEFORE UPDATE
    ON nomen.authorizations
    FOR EACH ROW
    WHEN (NEW.updated_at IS NULL)
EXECUTE FUNCTION nomen.set_updated_at();