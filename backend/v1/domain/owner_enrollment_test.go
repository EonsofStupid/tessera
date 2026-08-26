package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ownerEnrollmentDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

var ownerCredential = OwnerCredential{ID: []byte("credential-id"), PublicKey: []byte("cose-public-key"), AAGUID: make([]byte, 16), AttestationType: "none", Transports: []string{"internal"}}

func TestOwnerEnrollmentRequiresPasskeyThenRecovery(t *testing.T) {
	now := time.Date(2026, time.August, 22, 3, 0, 0, 0, time.UTC)
	enrollment, err := BeginOwnerEnrollment(nil, BeginOwnerEnrollmentInput{
		InstanceID: "instance-1", CeremonyID: "ceremony-1", OwnerID: "owner-1", OwnerUsername: "owner@example.test", OwnerDisplayName: "Owner",
		ChallengeDigest: ownerEnrollmentDigest, IdempotencyKeyDigest: ownerEnrollmentDigest,
		RequestDigest: ownerEnrollmentDigest, Now: now, ExpiresAt: now.Add(10 * time.Minute),
	})
	require.NoError(t, err)
	assert.Equal(t, OwnerEnrollmentPasskeyPending, enrollment.State)

	enrollment, err = RecordOwnerPasskey(*enrollment, "webauthn-credential-1", ownerCredential, ownerEnrollmentDigest, now.Add(time.Minute))
	require.NoError(t, err)
	assert.Equal(t, OwnerEnrollmentRecoveryPending, enrollment.State)

	enrollment, err = ConfirmOwnerRecovery(*enrollment, ownerEnrollmentDigest, now.Add(2*time.Minute))
	require.NoError(t, err)
	assert.Equal(t, OwnerEnrollmentComplete, enrollment.State)
	assert.Equal(t, uint64(3), enrollment.Revision)
	require.NoError(t, enrollment.Validate())
}

func TestOwnerEnrollmentIdempotencyRefusesPayloadReuse(t *testing.T) {
	now := time.Now().UTC()
	input := BeginOwnerEnrollmentInput{InstanceID: "instance-1", CeremonyID: "ceremony-1", OwnerID: "owner-1", OwnerUsername: "owner@example.test", OwnerDisplayName: "Owner", ChallengeDigest: ownerEnrollmentDigest, IdempotencyKeyDigest: ownerEnrollmentDigest, RequestDigest: ownerEnrollmentDigest, Now: now, ExpiresAt: now.Add(time.Minute)}
	current, err := BeginOwnerEnrollment(nil, input)
	require.NoError(t, err)

	retried, err := BeginOwnerEnrollment(current, input)
	require.NoError(t, err)
	assert.Equal(t, current, retried)

	input.RequestDigest = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	_, err = BeginOwnerEnrollment(current, input)
	var refusal *OwnerEnrollmentError
	require.ErrorAs(t, err, &refusal)
	assert.Equal(t, OwnerEnrollmentIdempotencyReuse, refusal.Reason)
}

func TestOwnerEnrollmentRejectsExpiredAndOutOfOrderTransitions(t *testing.T) {
	now := time.Now().UTC()
	current, err := BeginOwnerEnrollment(nil, BeginOwnerEnrollmentInput{InstanceID: "instance-1", CeremonyID: "ceremony-1", OwnerID: "owner-1", OwnerUsername: "owner@example.test", OwnerDisplayName: "Owner", ChallengeDigest: ownerEnrollmentDigest, IdempotencyKeyDigest: ownerEnrollmentDigest, RequestDigest: ownerEnrollmentDigest, Now: now, ExpiresAt: now.Add(time.Minute)})
	require.NoError(t, err)

	_, err = RecordOwnerPasskey(*current, "credential", ownerCredential, ownerEnrollmentDigest, now.Add(2*time.Minute))
	require.ErrorContains(t, err, "ceremony_expired")
	_, err = ConfirmOwnerRecovery(*current, ownerEnrollmentDigest, now)
	require.ErrorContains(t, err, "transition_out_of_order")
}

func TestOwnerEnrollmentViewExcludesCeremonyEvidence(t *testing.T) {
	now := time.Now().UTC()
	pending := BuildOwnerEnrollmentView(nil, now)
	require.NoError(t, pending.Validate())
	assert.Equal(t, OwnerEnrollmentPending, pending.State)
	assert.Zero(t, pending.Revision)

	enrollment, err := BeginOwnerEnrollment(nil, BeginOwnerEnrollmentInput{InstanceID: "instance-1", CeremonyID: "ceremony-1", OwnerID: "owner-1", OwnerUsername: "owner@example.test", OwnerDisplayName: "Owner", ChallengeDigest: ownerEnrollmentDigest, IdempotencyKeyDigest: ownerEnrollmentDigest, RequestDigest: ownerEnrollmentDigest, Now: now, ExpiresAt: now.Add(time.Minute)})
	require.NoError(t, err)
	view := BuildOwnerEnrollmentView(enrollment, now)
	require.NoError(t, view.Validate())
	assert.Equal(t, "ceremony-1", view.CeremonyID)
	assert.False(t, view.PasskeyEnrolled)
}
