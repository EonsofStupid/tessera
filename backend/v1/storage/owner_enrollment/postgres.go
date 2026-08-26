// Package owner_enrollment persists the resumable first-owner ceremony in the
// Nomen schema. It stores only digests and verified credential references;
// bootstrap authority and raw WebAuthn challenge values never enter this table.
package owner_enrollment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/shippinAI/nomen/backend/v3/storage/database"
)

type Repository struct {
	pool database.Pool
}

func NewRepository(pool database.Pool) *Repository {
	return &Repository{pool: pool}
}

var _ domain.OwnerEnrollmentRepository = (*Repository)(nil)

const selectEnrollment = `
SELECT state, ceremony_id, owner_id, owner_username, owner_display_name,
       challenge_digest, credential_reference,
       recovery_artifact_digest, idempotency_key_digest, request_digest,
       expires_at, created_at, updated_at, completed_at, revision,
       credential_id, credential_public_key, credential_sign_count,
       credential_aaguid, credential_attestation_type, credential_transports,
       credential_flags
FROM nomen_product.nomen_owner_enrollments
WHERE instance_id = $1`

func (r *Repository) Get(ctx context.Context, instanceID string) (enrollment *domain.OwnerEnrollment, err error) {
	if r.pool == nil || instanceID == "" {
		return nil, fmt.Errorf("owner enrollment read requires a pool and instance")
	}
	tx, err := r.pool.Begin(ctx, &database.TransactionOptions{IsolationLevel: database.IsolationLevelReadCommitted, AccessMode: database.AccessModeReadOnly})
	if err != nil {
		return nil, fmt.Errorf("begin owner enrollment read: %w", err)
	}
	defer func() { err = tx.End(ctx, err) }()
	if _, err = tx.Exec(ctx, `SELECT set_config('nomen_product.instance_id', $1, true)`, instanceID); err != nil {
		return nil, fmt.Errorf("set owner enrollment instance context: %w", err)
	}
	enrollment = &domain.OwnerEnrollment{InstanceID: instanceID}
	var state string
	var credentialID, credentialPublicKey, credentialAAGUID []byte
	var credentialSignCount sql.NullInt64
	var credentialAttestation sql.NullString
	var credentialTransports []string
	var credentialFlags sql.NullInt16
	err = tx.QueryRow(ctx, selectEnrollment, instanceID).Scan(
		&state, &enrollment.CeremonyID, &enrollment.OwnerID,
		&enrollment.OwnerUsername, &enrollment.OwnerDisplayName,
		&enrollment.ChallengeDigest, &enrollment.CredentialReference,
		&enrollment.RecoveryArtifactDigest, &enrollment.IdempotencyKeyDigest,
		&enrollment.RequestDigest, &enrollment.ExpiresAt, &enrollment.CreatedAt,
		&enrollment.UpdatedAt, &enrollment.CompletedAt, &enrollment.Revision,
		&credentialID, &credentialPublicKey, &credentialSignCount,
		&credentialAAGUID, &credentialAttestation, &credentialTransports,
		&credentialFlags,
	)
	if err != nil {
		if errors.Is(err, &database.NoRowFoundError{}) {
			return nil, nil
		}
		return nil, fmt.Errorf("read owner enrollment: %w", err)
	}
	enrollment.State = domain.OwnerEnrollmentState(state)
	if credentialSignCount.Valid && credentialAttestation.Valid && credentialFlags.Valid {
		enrollment.Credential = &domain.OwnerCredential{
			ID: credentialID, PublicKey: credentialPublicKey,
			SignCount: uint32(credentialSignCount.Int64), AAGUID: credentialAAGUID,
			AttestationType: credentialAttestation.String,
			Transports:      credentialTransports, Flags: byte(credentialFlags.Int16),
		}
	}
	if err = enrollment.Validate(); err != nil {
		return nil, fmt.Errorf("validate stored owner enrollment: %w", err)
	}
	return enrollment, nil
}

func (r *Repository) Save(ctx context.Context, enrollment *domain.OwnerEnrollment, expectedRevision uint64) (err error) {
	if r.pool == nil || enrollment == nil {
		return fmt.Errorf("owner enrollment write requires a pool and enrollment")
	}
	if err := enrollment.Validate(); err != nil {
		return err
	}
	if expectedRevision == 0 && enrollment.Revision != 1 || expectedRevision > 0 && enrollment.Revision != expectedRevision+1 {
		return &domain.OwnerEnrollmentError{Reason: domain.OwnerEnrollmentRevisionConflict, Field: "revision", Detail: "new revision must advance the expected revision exactly once"}
	}
	tx, err := r.pool.Begin(ctx, &database.TransactionOptions{IsolationLevel: database.IsolationLevelSerializable, AccessMode: database.AccessModeReadWrite})
	if err != nil {
		return fmt.Errorf("begin owner enrollment write: %w", err)
	}
	defer func() { err = tx.End(ctx, err) }()
	if _, err = tx.Exec(ctx, `SELECT set_config('nomen_product.instance_id', $1, true)`, enrollment.InstanceID); err != nil {
		return fmt.Errorf("set owner enrollment instance context: %w", err)
	}
	var affected int64
	if expectedRevision == 0 {
		affected, err = tx.Exec(ctx, `
INSERT INTO nomen_product.nomen_owner_enrollments (
    instance_id, state, ceremony_id, owner_id, owner_username,
    owner_display_name, challenge_digest,
    credential_reference, recovery_artifact_digest, idempotency_key_digest,
    request_digest, expires_at, created_at, updated_at, completed_at, revision
    , credential_id, credential_public_key, credential_sign_count,
    credential_aaguid, credential_attestation_type, credential_transports,
    credential_flags
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
          $15, $16, $17, $18, $19, $20, $21, $22, $23)
ON CONFLICT (instance_id) DO NOTHING`, enrollmentValues(enrollment)...)
	} else {
		values := enrollmentValues(enrollment)
		values = append(values, expectedRevision)
		affected, err = tx.Exec(ctx, `
UPDATE nomen_product.nomen_owner_enrollments SET
    state = $2, ceremony_id = $3, owner_id = $4, owner_username = $5,
    owner_display_name = $6, challenge_digest = $7,
    credential_reference = $8, recovery_artifact_digest = $9,
    idempotency_key_digest = $10, request_digest = $11, expires_at = $12,
    created_at = $13, updated_at = $14, completed_at = $15, revision = $16,
    credential_id = $17, credential_public_key = $18,
    credential_sign_count = $19, credential_aaguid = $20,
    credential_attestation_type = $21, credential_transports = $22,
    credential_flags = $23
WHERE instance_id = $1 AND revision = $24`, values...)
	}
	if err != nil {
		return fmt.Errorf("write owner enrollment: %w", err)
	}
	if affected != 1 {
		return &domain.OwnerEnrollmentError{Reason: domain.OwnerEnrollmentRevisionConflict, Field: "revision", Detail: "owner enrollment changed after it was read"}
	}
	return nil
}

func enrollmentValues(enrollment *domain.OwnerEnrollment) []any {
	var credentialID, credentialPublicKey, credentialAAGUID any
	var credentialSignCount, credentialAttestation, credentialTransports, credentialFlags any
	if enrollment.Credential != nil {
		credentialID = enrollment.Credential.ID
		credentialPublicKey = enrollment.Credential.PublicKey
		credentialSignCount = int64(enrollment.Credential.SignCount)
		credentialAAGUID = enrollment.Credential.AAGUID
		credentialAttestation = enrollment.Credential.AttestationType
		credentialTransports = enrollment.Credential.Transports
		credentialFlags = int16(enrollment.Credential.Flags)
	}
	return []any{
		enrollment.InstanceID, string(enrollment.State), enrollment.CeremonyID,
		enrollment.OwnerID, enrollment.OwnerUsername, enrollment.OwnerDisplayName,
		enrollment.ChallengeDigest, enrollment.CredentialReference,
		enrollment.RecoveryArtifactDigest, enrollment.IdempotencyKeyDigest,
		enrollment.RequestDigest, enrollment.ExpiresAt, enrollment.CreatedAt,
		enrollment.UpdatedAt, enrollment.CompletedAt, enrollment.Revision,
		credentialID, credentialPublicKey, credentialSignCount, credentialAAGUID,
		credentialAttestation, credentialTransports, credentialFlags,
	}
}
