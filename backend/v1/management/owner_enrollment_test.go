package management

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ownerEnrollmentRepositoryStub struct {
	enrollment *domain.OwnerEnrollment
	err        error
}

type ownerEnrollmentMemoryRepository struct {
	enrollment *domain.OwnerEnrollment
	saves      int
}

func (s *ownerEnrollmentMemoryRepository) Get(context.Context, string) (*domain.OwnerEnrollment, error) {
	if s.enrollment == nil {
		return nil, nil
	}
	cloned := *s.enrollment
	return &cloned, nil
}

func (s *ownerEnrollmentMemoryRepository) Save(_ context.Context, enrollment *domain.OwnerEnrollment, _ uint64) error {
	cloned := *enrollment
	s.enrollment = &cloned
	s.saves++
	return nil
}

func (s ownerEnrollmentRepositoryStub) Get(context.Context, string) (*domain.OwnerEnrollment, error) {
	return s.enrollment, s.err
}

func (s ownerEnrollmentRepositoryStub) Save(context.Context, *domain.OwnerEnrollment, uint64) error {
	return s.err
}

func TestOwnerEnrollmentServiceReturnsPendingWithoutInventingCeremony(t *testing.T) {
	now := time.Date(2026, time.August, 22, 4, 0, 0, 0, time.UTC)
	view, err := NewOwnerEnrollmentService(ownerEnrollmentRepositoryStub{}, func() time.Time { return now }).Get(context.Background(), "instance-1")
	require.NoError(t, err)
	assert.Equal(t, domain.OwnerEnrollmentPending, view.State)
	assert.Empty(t, view.CeremonyID)
	assert.Equal(t, now, view.ObservedAt)
}

func TestOwnerEnrollmentBeginUsesRealWebAuthnOptionsAndIdempotentChallenge(t *testing.T) {
	now := time.Date(2026, time.August, 22, 5, 0, 0, 0, time.UTC)
	repository := &ownerEnrollmentMemoryRepository{}
	service := NewOwnerEnrollmentService(repository, func() time.Time { return now }).WithBootstrapAuthority("0123456789abcdef0123456789abcdef")
	request := BeginOwnerEnrollmentRequest{OwnerID: "owner-1", Username: "jesse@example.test", DisplayName: "Jesse Hall"}

	first, err := service.Begin(context.Background(), "instance-1", "https://identity.example.test", "0123456789abcdef0123456789abcdef", "idempotency-key-0001", request)
	require.NoError(t, err)
	assert.True(t, first.Created)
	assert.Equal(t, "identity.example.test", first.PublicKey.RelyingParty.ID)
	assert.NotNil(t, first.PublicKey.User.ID)
	assert.NotEmpty(t, first.PublicKey.Challenge)
	assert.Equal(t, protocol.VerificationRequired, first.PublicKey.AuthenticatorSelection.UserVerification)
	assert.Equal(t, 1, repository.saves)
	assert.NotContains(t, repository.enrollment.ChallengeDigest, base64.RawURLEncoding.EncodeToString(first.PublicKey.Challenge))

	retried, err := service.Begin(context.Background(), "instance-1", "https://identity.example.test", "0123456789abcdef0123456789abcdef", "idempotency-key-0001", request)
	require.NoError(t, err)
	assert.False(t, retried.Created)
	assert.Equal(t, first.PublicKey.Challenge, retried.PublicKey.Challenge)
	assert.Equal(t, 1, repository.saves)

	resumed, err := service.Begin(context.Background(), "instance-1", "https://identity.example.test", "0123456789abcdef0123456789abcdef", "idempotency-key-0002", BeginOwnerEnrollmentRequest{OwnerID: "ignored-owner", Username: "ignored@example.test", DisplayName: "Ignored"})
	require.NoError(t, err)
	assert.False(t, resumed.Created)
	assert.Equal(t, first.Enrollment.OwnerID, resumed.Enrollment.OwnerID)
	assert.Equal(t, first.PublicKey.Challenge, resumed.PublicKey.Challenge)
	assert.Equal(t, 1, repository.saves)
}

func TestOwnerEnrollmentBeginRejectsWrongBootstrapAuthority(t *testing.T) {
	service := NewOwnerEnrollmentService(&ownerEnrollmentMemoryRepository{}, time.Now).WithBootstrapAuthority("0123456789abcdef0123456789abcdef")
	_, err := service.Begin(context.Background(), "instance-1", "https://identity.example.test", "wrong-authority-that-is-at-least-32-bytes", "idempotency-key-0001", BeginOwnerEnrollmentRequest{OwnerID: "owner-1", Username: "owner@example.test", DisplayName: "Owner"})
	var refusal *domain.OwnerEnrollmentError
	require.ErrorAs(t, err, &refusal)
	assert.Equal(t, "bootstrap_authority", refusal.Field)
}
