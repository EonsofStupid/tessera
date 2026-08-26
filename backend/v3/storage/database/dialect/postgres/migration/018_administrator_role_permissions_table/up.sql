CREATE TABLE nomen.administrator_role_permissions(
    instance_id TEXT NOT NULL
    , role_name TEXT NOT NULL CHECK (role_name <> '')
    , permission TEXT NOT NULL CHECK (permission <> '')

    , PRIMARY KEY (instance_id, permission, role_name)
    , FOREIGN KEY (instance_id) REFERENCES nomen.instances(id) ON DELETE CASCADE
);
