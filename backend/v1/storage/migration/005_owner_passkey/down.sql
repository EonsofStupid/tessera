ALTER TABLE nomen_product.nomen_owner_enrollments
    DROP CONSTRAINT nomen_owner_enrollment_state_evidence,
    DROP CONSTRAINT nomen_owner_enrollment_owner_identity,
    DROP CONSTRAINT nomen_owner_enrollment_credential_material;

ALTER TABLE nomen_product.nomen_owner_enrollments
    ADD CONSTRAINT nomen_owner_enrollments_check CHECK (
        (state = 'passkey_pending' AND credential_reference = '' AND recovery_artifact_digest = '' AND completed_at IS NULL)
        OR (state = 'recovery_pending' AND credential_reference <> '' AND recovery_artifact_digest = '' AND completed_at IS NULL)
        OR (state = 'complete' AND credential_reference <> '' AND recovery_artifact_digest ~ '^sha256:[0-9a-f]{64}$' AND completed_at IS NOT NULL)
    );

ALTER TABLE nomen_product.nomen_owner_enrollments
    DROP COLUMN owner_username,
    DROP COLUMN owner_display_name,
    DROP COLUMN credential_id,
    DROP COLUMN credential_public_key,
    DROP COLUMN credential_sign_count,
    DROP COLUMN credential_aaguid,
    DROP COLUMN credential_attestation_type,
    DROP COLUMN credential_transports,
    DROP COLUMN credential_flags;
