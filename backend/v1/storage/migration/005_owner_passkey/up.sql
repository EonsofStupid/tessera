ALTER TABLE nomen_product.nomen_owner_enrollments
    ADD COLUMN owner_username TEXT,
    ADD COLUMN owner_display_name TEXT,
    ADD COLUMN credential_id BYTEA,
    ADD COLUMN credential_public_key BYTEA,
    ADD COLUMN credential_sign_count BIGINT,
    ADD COLUMN credential_aaguid BYTEA,
    ADD COLUMN credential_attestation_type TEXT,
    ADD COLUMN credential_transports TEXT[],
    ADD COLUMN credential_flags SMALLINT;

UPDATE nomen_product.nomen_owner_enrollments
SET owner_username = owner_id, owner_display_name = owner_id
WHERE owner_username IS NULL OR owner_display_name IS NULL;

ALTER TABLE nomen_product.nomen_owner_enrollments
    ALTER COLUMN owner_username SET NOT NULL,
    ALTER COLUMN owner_display_name SET NOT NULL;

ALTER TABLE nomen_product.nomen_owner_enrollments
    ADD CONSTRAINT nomen_owner_enrollment_owner_identity CHECK (
        length(owner_username) BETWEEN 1 AND 320
        AND length(owner_display_name) BETWEEN 1 AND 200
    );

ALTER TABLE nomen_product.nomen_owner_enrollments
    ADD CONSTRAINT nomen_owner_enrollment_credential_material CHECK (
        (state = 'passkey_pending'
            AND credential_id IS NULL
            AND credential_public_key IS NULL
            AND credential_sign_count IS NULL
            AND credential_aaguid IS NULL
            AND credential_attestation_type IS NULL
            AND credential_transports IS NULL
            AND credential_flags IS NULL)
        OR
        (state IN ('recovery_pending', 'complete')
            AND octet_length(credential_id) BETWEEN 1 AND 1024
            AND octet_length(credential_public_key) BETWEEN 1 AND 8192
            AND credential_sign_count BETWEEN 0 AND 4294967295
            AND octet_length(credential_aaguid) <= 64
            AND length(credential_attestation_type) <= 64
            AND cardinality(credential_transports) <= 16
            AND credential_flags BETWEEN 0 AND 255)
    );

ALTER TABLE nomen_product.nomen_owner_enrollments
    DROP CONSTRAINT nomen_owner_enrollments_check;

ALTER TABLE nomen_product.nomen_owner_enrollments
    ADD CONSTRAINT nomen_owner_enrollment_state_evidence CHECK (
        (state = 'passkey_pending' AND credential_reference = '' AND recovery_artifact_digest = '' AND completed_at IS NULL)
        OR (state = 'recovery_pending' AND credential_reference <> '' AND recovery_artifact_digest ~ '^sha256:[0-9a-f]{64}$' AND completed_at IS NULL)
        OR (state = 'complete' AND credential_reference <> '' AND recovery_artifact_digest ~ '^sha256:[0-9a-f]{64}$' AND completed_at IS NOT NULL)
    );
