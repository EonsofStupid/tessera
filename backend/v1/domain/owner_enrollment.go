package domain

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type OwnerEnrollmentState string

const (
	OwnerEnrollmentPending         OwnerEnrollmentState = "pending"
	OwnerEnrollmentPasskeyPending  OwnerEnrollmentState = "passkey_pending"
	OwnerEnrollmentRecoveryPending OwnerEnrollmentState = "recovery_pending"
	OwnerEnrollmentComplete        OwnerEnrollmentState = "complete"
)

func (s OwnerEnrollmentState) Valid() bool {
	return s == OwnerEnrollmentPending || s == OwnerEnrollmentPasskeyPending || s == OwnerEnrollmentRecoveryPending || s == OwnerEnrollmentComplete
}

type OwnerEnrollment struct {
	InstanceID             string               `json:"-"`
	State                  OwnerEnrollmentState `json:"state"`
	CeremonyID             string               `json:"ceremony_id"`
	OwnerID                string               `json:"owner_id"`
	OwnerUsername          string               `json:"-"`
	OwnerDisplayName       string               `json:"-"`
	ChallengeDigest        string               `json:"-"`
	CredentialReference    string               `json:"credential_reference,omitempty"`
	Credential             *OwnerCredential     `json:"-"`
	RecoveryArtifactDigest string               `json:"recovery_artifact_digest,omitempty"`
	IdempotencyKeyDigest   string               `json:"-"`
	RequestDigest          string               `json:"-"`
	ExpiresAt              time.Time            `json:"expires_at"`
	CreatedAt              time.Time            `json:"created_at"`
	UpdatedAt              time.Time            `json:"updated_at"`
	CompletedAt            *time.Time           `json:"completed_at,omitempty"`
	Revision               uint64               `json:"revision"`
}

// OwnerCredential is WebAuthn public material. The authenticator keeps the
// private key; Nomen persists only what is required to verify later proofs.
type OwnerCredential struct {
	ID              []byte
	PublicKey       []byte
	SignCount       uint32
	AAGUID          []byte
	AttestationType string
	Transports      []string
	Flags           byte
}

func (c OwnerCredential) Validate() error {
	if len(c.ID) == 0 || len(c.ID) > 1024 || len(c.PublicKey) == 0 || len(c.PublicKey) > 8192 || len(c.AAGUID) > 64 || len(c.AttestationType) > 64 || len(c.Transports) > 16 {
		return ownerEnrollmentError(OwnerEnrollmentInvalid, "credential", "verified WebAuthn public material is incomplete or exceeds its bound")
	}
	for _, transport := range c.Transports {
		if strings.TrimSpace(transport) == "" || len(transport) > 32 {
			return ownerEnrollmentError(OwnerEnrollmentInvalid, "credential.transports", "credential transport is invalid")
		}
	}
	return nil
}

const OwnerEnrollmentViewSchemaVersion uint32 = 1

type OwnerEnrollmentView struct {
	SchemaVersion     uint32               `json:"schema_version"`
	ResourceRevision  string               `json:"resource_revision"`
	ObservedAt        time.Time            `json:"observed_at"`
	State             OwnerEnrollmentState `json:"state"`
	CeremonyID        string               `json:"ceremony_id,omitempty"`
	OwnerID           string               `json:"owner_id,omitempty"`
	PasskeyEnrolled   bool                 `json:"passkey_enrolled"`
	RecoveryConfirmed bool                 `json:"recovery_confirmed"`
	ExpiresAt         *time.Time           `json:"expires_at,omitempty"`
	Revision          uint64               `json:"revision"`
}

func BuildOwnerEnrollmentView(enrollment *OwnerEnrollment, observedAt time.Time) OwnerEnrollmentView {
	view := OwnerEnrollmentView{SchemaVersion: OwnerEnrollmentViewSchemaVersion, ObservedAt: observedAt.UTC(), State: OwnerEnrollmentPending}
	if enrollment != nil {
		view.State = enrollment.State
		view.CeremonyID = enrollment.CeremonyID
		view.OwnerID = enrollment.OwnerID
		view.PasskeyEnrolled = enrollment.State == OwnerEnrollmentRecoveryPending || enrollment.State == OwnerEnrollmentComplete
		view.RecoveryConfirmed = enrollment.State == OwnerEnrollmentComplete
		view.ExpiresAt = &enrollment.ExpiresAt
		view.Revision = enrollment.Revision
	}
	view.ResourceRevision = digest(struct {
		State             OwnerEnrollmentState
		CeremonyID        string
		OwnerID           string
		PasskeyEnrolled   bool
		RecoveryConfirmed bool
		ExpiresAt         *time.Time
		Revision          uint64
	}{view.State, view.CeremonyID, view.OwnerID, view.PasskeyEnrolled, view.RecoveryConfirmed, view.ExpiresAt, view.Revision})
	return view
}

func (v OwnerEnrollmentView) Validate() error {
	if v.SchemaVersion != OwnerEnrollmentViewSchemaVersion || !validPlanDigest(v.ResourceRevision) || v.ObservedAt.IsZero() || !v.State.Valid() {
		return ownerEnrollmentError(OwnerEnrollmentInvalid, "view", "owner-enrollment view identity is incomplete")
	}
	if v.State == OwnerEnrollmentPending {
		if v.CeremonyID != "" || v.OwnerID != "" || v.PasskeyEnrolled || v.RecoveryConfirmed || v.ExpiresAt != nil || v.Revision != 0 {
			return ownerEnrollmentError(OwnerEnrollmentInvalid, "view", "pending view cannot expose ceremony state")
		}
		return nil
	}
	if v.CeremonyID == "" || v.OwnerID == "" || v.ExpiresAt == nil || v.ExpiresAt.IsZero() || v.Revision == 0 {
		return ownerEnrollmentError(OwnerEnrollmentInvalid, "view", "active view requires ceremony identity and revision")
	}
	if v.PasskeyEnrolled != (v.State == OwnerEnrollmentRecoveryPending || v.State == OwnerEnrollmentComplete) || v.RecoveryConfirmed != (v.State == OwnerEnrollmentComplete) {
		return ownerEnrollmentError(OwnerEnrollmentInvalid, "view", "view completion flags do not match state")
	}
	return nil
}

type OwnerEnrollmentRefusal string

const (
	OwnerEnrollmentInvalid          OwnerEnrollmentRefusal = "invalid_enrollment"
	OwnerEnrollmentExpired          OwnerEnrollmentRefusal = "ceremony_expired"
	OwnerEnrollmentOutOfOrder       OwnerEnrollmentRefusal = "transition_out_of_order"
	OwnerEnrollmentReplay           OwnerEnrollmentRefusal = "ceremony_replayed"
	OwnerEnrollmentIdempotencyReuse OwnerEnrollmentRefusal = "idempotency_key_reused"
	OwnerEnrollmentRevisionConflict OwnerEnrollmentRefusal = "revision_conflict"
)

type OwnerEnrollmentError struct {
	Reason OwnerEnrollmentRefusal `json:"reason"`
	Field  string                 `json:"field,omitempty"`
	Detail string                 `json:"detail"`
}

func (e *OwnerEnrollmentError) Error() string {
	return fmt.Sprintf("owner enrollment: %s (%s): %s", e.Reason, e.Field, e.Detail)
}

type BeginOwnerEnrollmentInput struct {
	InstanceID           string
	CeremonyID           string
	OwnerID              string
	OwnerUsername        string
	OwnerDisplayName     string
	ChallengeDigest      string
	IdempotencyKeyDigest string
	RequestDigest        string
	ExpiresAt            time.Time
	Now                  time.Time
}

func BeginOwnerEnrollment(current *OwnerEnrollment, input BeginOwnerEnrollmentInput) (*OwnerEnrollment, error) {
	if strings.TrimSpace(input.InstanceID) == "" || strings.TrimSpace(input.CeremonyID) == "" || strings.TrimSpace(input.OwnerID) == "" || strings.TrimSpace(input.OwnerUsername) == "" || strings.TrimSpace(input.OwnerDisplayName) == "" {
		return nil, ownerEnrollmentError(OwnerEnrollmentInvalid, "identity", "instance, ceremony and owner are required")
	}
	if len(input.OwnerID) > 200 || len(input.OwnerUsername) > 320 || len(input.OwnerDisplayName) > 200 {
		return nil, ownerEnrollmentError(OwnerEnrollmentInvalid, "identity", "owner identity exceeds its bound")
	}
	if !validPlanDigest(input.ChallengeDigest) || !validPlanDigest(input.IdempotencyKeyDigest) || !validPlanDigest(input.RequestDigest) {
		return nil, ownerEnrollmentError(OwnerEnrollmentInvalid, "digest", "challenge, idempotency and request digests must be lowercase sha256 values")
	}
	if input.Now.IsZero() || !input.ExpiresAt.After(input.Now) {
		return nil, ownerEnrollmentError(OwnerEnrollmentInvalid, "expires_at", "ceremony expiry must be after creation")
	}
	if current != nil {
		if err := current.Validate(); err != nil {
			return nil, err
		}
		if current.InstanceID != input.InstanceID {
			return nil, ownerEnrollmentError(OwnerEnrollmentReplay, "instance_id", "ceremony belongs to another instance")
		}
		if current.IdempotencyKeyDigest == input.IdempotencyKeyDigest {
			if current.RequestDigest != input.RequestDigest {
				return nil, ownerEnrollmentError(OwnerEnrollmentIdempotencyReuse, "idempotency_key", "key was already used for a different owner-enrollment request")
			}
			cloned := *current
			return &cloned, nil
		}
		return nil, ownerEnrollmentError(OwnerEnrollmentOutOfOrder, "state", "an owner-enrollment ceremony already exists")
	}
	now := input.Now.UTC()
	return &OwnerEnrollment{
		InstanceID: input.InstanceID, State: OwnerEnrollmentPasskeyPending,
		CeremonyID: input.CeremonyID, OwnerID: input.OwnerID,
		OwnerUsername: input.OwnerUsername, OwnerDisplayName: input.OwnerDisplayName,
		ChallengeDigest: input.ChallengeDigest, IdempotencyKeyDigest: input.IdempotencyKeyDigest,
		RequestDigest: input.RequestDigest, ExpiresAt: input.ExpiresAt.UTC(),
		CreatedAt: now, UpdatedAt: now, Revision: 1,
	}, nil
}

func RecordOwnerPasskey(current OwnerEnrollment, credentialReference string, credential OwnerCredential, recoveryArtifactDigest string, now time.Time) (*OwnerEnrollment, error) {
	if err := current.Validate(); err != nil {
		return nil, err
	}
	if current.State != OwnerEnrollmentPasskeyPending {
		return nil, ownerEnrollmentError(OwnerEnrollmentOutOfOrder, "state", "passkey registration is not expected in the current state")
	}
	if now.IsZero() || !now.Before(current.ExpiresAt) {
		return nil, ownerEnrollmentError(OwnerEnrollmentExpired, "expires_at", "the passkey ceremony has expired")
	}
	if strings.TrimSpace(credentialReference) == "" {
		return nil, ownerEnrollmentError(OwnerEnrollmentInvalid, "credential_reference", "a verified credential reference is required")
	}
	if err := credential.Validate(); err != nil {
		return nil, err
	}
	if !validPlanDigest(recoveryArtifactDigest) {
		return nil, ownerEnrollmentError(OwnerEnrollmentInvalid, "recovery_artifact_digest", "a generated recovery-artifact digest is required")
	}
	next := current
	next.State = OwnerEnrollmentRecoveryPending
	next.CredentialReference = credentialReference
	next.Credential = &credential
	next.RecoveryArtifactDigest = recoveryArtifactDigest
	next.UpdatedAt = now.UTC()
	next.Revision++
	return &next, nil
}

func ConfirmOwnerRecovery(current OwnerEnrollment, artifactDigest string, now time.Time) (*OwnerEnrollment, error) {
	if err := current.Validate(); err != nil {
		return nil, err
	}
	if current.State == OwnerEnrollmentComplete {
		if current.RecoveryArtifactDigest == artifactDigest {
			cloned := current
			return &cloned, nil
		}
		return nil, ownerEnrollmentError(OwnerEnrollmentReplay, "recovery_artifact_digest", "completed enrollment cannot be rebound to another recovery artifact")
	}
	if current.State != OwnerEnrollmentRecoveryPending {
		return nil, ownerEnrollmentError(OwnerEnrollmentOutOfOrder, "state", "recovery confirmation requires a verified passkey")
	}
	if !validPlanDigest(artifactDigest) || now.IsZero() {
		return nil, ownerEnrollmentError(OwnerEnrollmentInvalid, "recovery_artifact_digest", "a lowercase sha256 artifact digest and observation time are required")
	}
	if current.RecoveryArtifactDigest != artifactDigest {
		return nil, ownerEnrollmentError(OwnerEnrollmentReplay, "recovery_artifact_digest", "recovery artifact does not match the generated enrollment artifact")
	}
	next := current
	next.State = OwnerEnrollmentComplete
	next.UpdatedAt = now.UTC()
	next.CompletedAt = &next.UpdatedAt
	next.Revision++
	return &next, nil
}

func (e OwnerEnrollment) Validate() error {
	if strings.TrimSpace(e.InstanceID) == "" || !e.State.Valid() || strings.TrimSpace(e.CeremonyID) == "" || strings.TrimSpace(e.OwnerID) == "" || strings.TrimSpace(e.OwnerUsername) == "" || strings.TrimSpace(e.OwnerDisplayName) == "" {
		return ownerEnrollmentError(OwnerEnrollmentInvalid, "identity", "persisted enrollment identity is incomplete")
	}
	if !validPlanDigest(e.ChallengeDigest) || !validPlanDigest(e.IdempotencyKeyDigest) || !validPlanDigest(e.RequestDigest) || e.ExpiresAt.IsZero() || e.CreatedAt.IsZero() || e.UpdatedAt.IsZero() || e.Revision == 0 {
		return ownerEnrollmentError(OwnerEnrollmentInvalid, "state", "persisted enrollment evidence is incomplete")
	}
	if (e.State == OwnerEnrollmentRecoveryPending || e.State == OwnerEnrollmentComplete) && strings.TrimSpace(e.CredentialReference) == "" {
		return ownerEnrollmentError(OwnerEnrollmentInvalid, "credential_reference", "verified passkey reference is missing")
	}
	if (e.State == OwnerEnrollmentRecoveryPending || e.State == OwnerEnrollmentComplete) && (e.Credential == nil || e.Credential.Validate() != nil || !validPlanDigest(e.RecoveryArtifactDigest)) {
		return ownerEnrollmentError(OwnerEnrollmentInvalid, "credential", "verified passkey and recovery evidence are missing")
	}
	if e.State == OwnerEnrollmentComplete && (e.CompletedAt == nil || e.CompletedAt.IsZero()) {
		return ownerEnrollmentError(OwnerEnrollmentInvalid, "recovery_artifact_digest", "completed enrollment requires recovery evidence")
	}
	if e.State == OwnerEnrollmentPasskeyPending && (e.Credential != nil || e.RecoveryArtifactDigest != "" || e.CompletedAt != nil) {
		return ownerEnrollmentError(OwnerEnrollmentInvalid, "completed_at", "incomplete enrollment cannot carry completion evidence")
	}
	if e.State == OwnerEnrollmentRecoveryPending && e.CompletedAt != nil {
		return ownerEnrollmentError(OwnerEnrollmentInvalid, "completed_at", "recovery-pending enrollment cannot be complete")
	}
	return nil
}

type OwnerEnrollmentRepository interface {
	Get(ctx context.Context, instanceID string) (*OwnerEnrollment, error)
	Save(ctx context.Context, enrollment *OwnerEnrollment, expectedRevision uint64) error
}

func ownerEnrollmentError(reason OwnerEnrollmentRefusal, field, detail string) error {
	return &OwnerEnrollmentError{Reason: reason, Field: field, Detail: detail}
}
