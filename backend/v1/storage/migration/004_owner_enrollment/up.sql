CREATE TABLE nomen_product.nomen_owner_enrollments (
    instance_id TEXT PRIMARY KEY,
    state TEXT NOT NULL CHECK (state IN ('passkey_pending', 'recovery_pending', 'complete')),
    ceremony_id TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    challenge_digest TEXT NOT NULL CHECK (challenge_digest ~ '^sha256:[0-9a-f]{64}$'),
    credential_reference TEXT NOT NULL DEFAULT '',
    recovery_artifact_digest TEXT NOT NULL DEFAULT '',
    idempotency_key_digest TEXT NOT NULL CHECK (idempotency_key_digest ~ '^sha256:[0-9a-f]{64}$'),
    request_digest TEXT NOT NULL CHECK (request_digest ~ '^sha256:[0-9a-f]{64}$'),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    revision BIGINT NOT NULL CHECK (revision > 0),
    CHECK ((state = 'passkey_pending' AND credential_reference = '' AND recovery_artifact_digest = '' AND completed_at IS NULL)
        OR (state = 'recovery_pending' AND credential_reference <> '' AND recovery_artifact_digest = '' AND completed_at IS NULL)
        OR (state = 'complete' AND credential_reference <> '' AND recovery_artifact_digest ~ '^sha256:[0-9a-f]{64}$' AND completed_at IS NOT NULL))
);

ALTER TABLE nomen_product.nomen_owner_enrollments ENABLE ROW LEVEL SECURITY;
ALTER TABLE nomen_product.nomen_owner_enrollments FORCE ROW LEVEL SECURITY;

CREATE POLICY nomen_owner_enrollments_instance_policy ON nomen_product.nomen_owner_enrollments
    USING (instance_id = current_setting('nomen_product.instance_id', true))
    WITH CHECK (instance_id = current_setting('nomen_product.instance_id', true));
