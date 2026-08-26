CREATE TYPE nomen.administrator_scope AS ENUM (
    'instance',
    'organization',
    'project',
    'project_grant'
);

CREATE TABLE nomen.administrators(
    instance_id TEXT NOT NULL
    , user_id TEXT NOT NULL
    , scope nomen.administrator_scope NOT NULL

    , created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    , updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()

    , organization_id TEXT
    , project_id TEXT
    , project_grant_id TEXT

    , id TEXT GENERATED ALWAYS AS (
        user_id || ':' ||
        CASE scope
            WHEN 'instance'::nomen.administrator_scope
                THEN 'instance:' || instance_id
            WHEN 'organization'::nomen.administrator_scope
                THEN 'organization:' || organization_id
            WHEN 'project'::nomen.administrator_scope
                THEN 'project:' || project_id
            WHEN 'project_grant'::nomen.administrator_scope
                THEN 'project_grant:' || project_grant_id
        END
    ) STORED

    , CONSTRAINT administrators_scope_alignment_chk CHECK (
        (scope = 'instance'
            AND organization_id IS NULL
            AND project_id IS NULL
            AND project_grant_id IS NULL)
        OR
        (scope = 'organization'
            AND organization_id IS NOT NULL
            AND project_id IS NULL
            AND project_grant_id IS NULL)
        OR
        (scope = 'project'
            AND organization_id IS NULL
            AND project_id IS NOT NULL
            AND project_grant_id IS NULL)
        OR
        (scope = 'project_grant'
            AND organization_id IS NULL
            AND project_id IS NULL
            AND project_grant_id IS NOT NULL)
    )

    , PRIMARY KEY (instance_id, id)
    , FOREIGN KEY (instance_id) REFERENCES nomen.instances(id) ON DELETE CASCADE
    , FOREIGN KEY (instance_id, user_id) REFERENCES nomen.users(instance_id, id) ON DELETE CASCADE
    , FOREIGN KEY (instance_id, organization_id) REFERENCES nomen.organizations(instance_id, id) ON DELETE CASCADE
    , FOREIGN KEY (instance_id, project_id) REFERENCES nomen.projects(instance_id, id) ON DELETE CASCADE
    , FOREIGN KEY (instance_id, project_grant_id) REFERENCES nomen.project_grants(instance_id, id) ON DELETE CASCADE
);

CREATE TABLE nomen.administrator_roles(
    instance_id TEXT NOT NULL
    , administrator_id TEXT NOT NULL
    , role_name TEXT NOT NULL CHECK (role_name <> '')

    , PRIMARY KEY (instance_id, administrator_id, role_name)
    , FOREIGN KEY (instance_id, administrator_id) REFERENCES nomen.administrators(instance_id, id) ON DELETE CASCADE
);

CREATE TRIGGER trigger_set_updated_at
    BEFORE UPDATE
    ON nomen.administrators
    FOR EACH ROW
    WHEN (NEW.updated_at IS NULL)
EXECUTE FUNCTION nomen.set_updated_at();
