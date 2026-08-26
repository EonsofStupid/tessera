package management

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-webauthn/webauthn/protocol"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"

	"github.com/shippinAI/nomen/backend/v1/domain"
)

const (
	ownerCeremonyLifetime  = 5 * time.Minute
	recoveryArtifactPrefix = "nomen-recovery-v1."
)

type BeginOwnerEnrollmentRequest struct {
	OwnerID     string `json:"owner_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type BeginOwnerEnrollmentResult struct {
	Enrollment domain.OwnerEnrollmentView                  `json:"enrollment"`
	PublicKey  protocol.PublicKeyCredentialCreationOptions `json:"publicKey"`
	Created    bool                                        `json:"-"`
}

type CompleteOwnerEnrollmentRequest struct {
	CeremonyID string          `json:"ceremony_id"`
	Credential json.RawMessage `json:"credential"`
}

type CompleteOwnerEnrollmentResult struct {
	Enrollment       domain.OwnerEnrollmentView `json:"enrollment"`
	RecoveryArtifact string                     `json:"recovery_artifact"`
}

type ConfirmOwnerRecoveryRequest struct {
	CeremonyID       string `json:"ceremony_id"`
	RecoveryArtifact string `json:"recovery_artifact"`
}

type OwnerEnrollmentService struct {
	repository      domain.OwnerEnrollmentRepository
	clock           Clock
	random          io.Reader
	bootstrapDigest [sha256.Size]byte
	bootstrapSet    bool
}

func NewOwnerEnrollmentService(repository domain.OwnerEnrollmentRepository, clock Clock) *OwnerEnrollmentService {
	if clock == nil {
		clock = time.Now
	}
	return &OwnerEnrollmentService{repository: repository, clock: clock, random: rand.Reader}
}

// WithBootstrapAuthority keeps only a process-memory digest of the runtime
// secret. The authority is deliberately not accepted when it is too short.
func (s *OwnerEnrollmentService) WithBootstrapAuthority(authority string) *OwnerEnrollmentService {
	if len(authority) >= 32 {
		s.bootstrapDigest = sha256.Sum256([]byte(authority))
		s.bootstrapSet = true
	}
	return s
}

func (s *OwnerEnrollmentService) Get(ctx context.Context, instanceID string) (domain.OwnerEnrollmentView, error) {
	enrollment, err := s.load(ctx, instanceID)
	if err != nil {
		return domain.OwnerEnrollmentView{}, err
	}
	view := domain.BuildOwnerEnrollmentView(enrollment, s.clock())
	if err := view.Validate(); err != nil {
		return domain.OwnerEnrollmentView{}, fmt.Errorf("assembled owner enrollment is invalid: %w", err)
	}
	return view, nil
}

func (s *OwnerEnrollmentService) GetWithBootstrap(ctx context.Context, instanceID, authority string) (domain.OwnerEnrollmentView, error) {
	enrollment, err := s.load(ctx, instanceID)
	if err != nil {
		return domain.OwnerEnrollmentView{}, err
	}
	if err := s.authorizeBootstrap(authority, enrollment); err != nil {
		return domain.OwnerEnrollmentView{}, err
	}
	view := domain.BuildOwnerEnrollmentView(enrollment, s.clock())
	if err := view.Validate(); err != nil {
		return domain.OwnerEnrollmentView{}, fmt.Errorf("assembled bootstrap owner enrollment is invalid: %w", err)
	}
	return view, nil
}

func (s *OwnerEnrollmentService) Begin(ctx context.Context, instanceID, issuer, authority, idempotencyKey string, request BeginOwnerEnrollmentRequest) (BeginOwnerEnrollmentResult, error) {
	current, err := s.load(ctx, instanceID)
	if err != nil {
		return BeginOwnerEnrollmentResult{}, err
	}
	if err := s.authorizeBootstrap(authority, current); err != nil {
		return BeginOwnerEnrollmentResult{}, err
	}
	request = normalizeOwnerRequest(request)
	if err := validateBeginRequest(request, idempotencyKey); err != nil {
		return BeginOwnerEnrollmentResult{}, err
	}

	now := s.clock().UTC()
	created := current == nil
	ceremonyID := ""
	expiresAt := now.Add(ownerCeremonyLifetime)
	if created {
		ceremonyID, err = s.randomID(24)
		if err != nil {
			return BeginOwnerEnrollmentResult{}, fmt.Errorf("generate ceremony identity: %w", err)
		}
	} else {
		if current.State != domain.OwnerEnrollmentPasskeyPending {
			return BeginOwnerEnrollmentResult{}, ownerRefusal(domain.OwnerEnrollmentOutOfOrder, "state", "passkey registration is no longer pending")
		}
		if !now.Before(current.ExpiresAt) {
			return BeginOwnerEnrollmentResult{}, ownerRefusal(domain.OwnerEnrollmentExpired, "expires_at", "owner-enrollment ceremony has expired")
		}
		ceremonyID = current.CeremonyID
		expiresAt = current.ExpiresAt
	}
	challenge := s.deriveChallenge(instanceID, ceremonyID)
	requestDigest := digestJSON(request)
	beginInput := domain.BeginOwnerEnrollmentInput{
		InstanceID: instanceID, CeremonyID: ceremonyID, OwnerID: request.OwnerID,
		OwnerUsername: request.Username, OwnerDisplayName: request.DisplayName,
		ChallengeDigest:      digestString(challenge),
		IdempotencyKeyDigest: digestString(idempotencyKey), RequestDigest: requestDigest,
		ExpiresAt: expiresAt, Now: now,
	}
	var enrollment *domain.OwnerEnrollment
	if current != nil {
		if current.IdempotencyKeyDigest == beginInput.IdempotencyKeyDigest && current.RequestDigest != beginInput.RequestDigest {
			return BeginOwnerEnrollmentResult{}, ownerRefusal(domain.OwnerEnrollmentIdempotencyReuse, "idempotency_key", "key was already used for a different owner-enrollment request")
		}
		enrollment = current
	} else {
		enrollment, err = domain.BeginOwnerEnrollment(nil, beginInput)
		if err != nil {
			return BeginOwnerEnrollmentResult{}, err
		}
	}
	if created {
		if err := s.repository.Save(ctx, enrollment, 0); err != nil {
			return BeginOwnerEnrollmentResult{}, fmt.Errorf("persist owner enrollment: %w", err)
		}
	}
	publicKey, err := ownerCreationOptions(issuer, *enrollment, challenge)
	if err != nil {
		return BeginOwnerEnrollmentResult{}, err
	}
	view := domain.BuildOwnerEnrollmentView(enrollment, now)
	return BeginOwnerEnrollmentResult{Enrollment: view, PublicKey: publicKey, Created: created}, nil
}

func (s *OwnerEnrollmentService) Complete(ctx context.Context, instanceID, issuer, authority string, request CompleteOwnerEnrollmentRequest) (CompleteOwnerEnrollmentResult, error) {
	current, err := s.load(ctx, instanceID)
	if err != nil {
		return CompleteOwnerEnrollmentResult{}, err
	}
	if err := s.authorizeBootstrap(authority, current); err != nil {
		return CompleteOwnerEnrollmentResult{}, err
	}
	if current == nil || request.CeremonyID == "" || request.CeremonyID != current.CeremonyID || len(request.Credential) == 0 {
		return CompleteOwnerEnrollmentResult{}, ownerRefusal(domain.OwnerEnrollmentReplay, "ceremony_id", "owner-enrollment ceremony does not match")
	}
	parsed, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(request.Credential))
	if err != nil {
		return CompleteOwnerEnrollmentResult{}, ownerRefusal(domain.OwnerEnrollmentInvalid, "credential", "credential response is not valid WebAuthn JSON")
	}
	challenge := parsed.Response.CollectedClientData.Challenge
	if !constantDigestEqual(current.ChallengeDigest, digestString(challenge)) {
		return CompleteOwnerEnrollmentResult{}, ownerRefusal(domain.OwnerEnrollmentReplay, "credential.challenge", "credential challenge does not match this ceremony")
	}
	server, _, err := ownerWebAuthnServer(issuer)
	if err != nil {
		return CompleteOwnerEnrollmentResult{}, err
	}
	credential, err := server.CreateCredential(
		ownerWebAuthnUserFromEnrollment(*current),
		gowebauthn.SessionData{
			Challenge: challenge, UserID: []byte(current.OwnerID),
			Expires:          current.ExpiresAt,
			UserVerification: protocol.VerificationRequired,
		},
		parsed,
	)
	if err != nil {
		return CompleteOwnerEnrollmentResult{}, ownerRefusal(domain.OwnerEnrollmentInvalid, "credential", "credential failed WebAuthn verification")
	}
	recoveryArtifact, err := s.recoveryArtifact()
	if err != nil {
		return CompleteOwnerEnrollmentResult{}, fmt.Errorf("generate recovery artifact: %w", err)
	}
	publicCredential := ownerCredentialFromWebAuthn(credential)
	referenceHash := sha256.Sum256(credential.ID)
	next, err := domain.RecordOwnerPasskey(
		*current,
		"webauthn:"+base64.RawURLEncoding.EncodeToString(referenceHash[:]),
		publicCredential,
		digestString(recoveryArtifact),
		s.clock().UTC(),
	)
	if err != nil {
		return CompleteOwnerEnrollmentResult{}, err
	}
	if err := s.repository.Save(ctx, next, current.Revision); err != nil {
		return CompleteOwnerEnrollmentResult{}, fmt.Errorf("persist verified owner credential: %w", err)
	}
	return CompleteOwnerEnrollmentResult{Enrollment: domain.BuildOwnerEnrollmentView(next, s.clock()), RecoveryArtifact: recoveryArtifact}, nil
}

func (s *OwnerEnrollmentService) ConfirmRecovery(ctx context.Context, instanceID, authority string, request ConfirmOwnerRecoveryRequest) (domain.OwnerEnrollmentView, error) {
	current, err := s.load(ctx, instanceID)
	if err != nil {
		return domain.OwnerEnrollmentView{}, err
	}
	if err := s.authorizeBootstrap(authority, current); err != nil {
		return domain.OwnerEnrollmentView{}, err
	}
	if current == nil || request.CeremonyID == "" || request.CeremonyID != current.CeremonyID {
		return domain.OwnerEnrollmentView{}, ownerRefusal(domain.OwnerEnrollmentReplay, "ceremony_id", "owner-enrollment ceremony does not match")
	}
	if !constantDigestEqual(current.RecoveryArtifactDigest, digestString(request.RecoveryArtifact)) {
		return domain.OwnerEnrollmentView{}, ownerRefusal(domain.OwnerEnrollmentReplay, "recovery_artifact", "recovery artifact does not match")
	}
	next, err := domain.ConfirmOwnerRecovery(*current, digestString(request.RecoveryArtifact), s.clock().UTC())
	if err != nil {
		return domain.OwnerEnrollmentView{}, err
	}
	if next.Revision != current.Revision {
		if err := s.repository.Save(ctx, next, current.Revision); err != nil {
			return domain.OwnerEnrollmentView{}, fmt.Errorf("persist owner recovery confirmation: %w", err)
		}
	}
	return domain.BuildOwnerEnrollmentView(next, s.clock()), nil
}

func (s *OwnerEnrollmentService) load(ctx context.Context, instanceID string) (*domain.OwnerEnrollment, error) {
	if s.repository == nil || s.clock == nil || s.random == nil || strings.TrimSpace(instanceID) == "" {
		return nil, fmt.Errorf("owner enrollment service is incomplete")
	}
	enrollment, err := s.repository.Get(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("owner enrollment unavailable: %w", err)
	}
	return enrollment, nil
}

func (s *OwnerEnrollmentService) authorizeBootstrap(authority string, current *domain.OwnerEnrollment) error {
	if current != nil && current.State == domain.OwnerEnrollmentComplete {
		return ownerRefusal(domain.OwnerEnrollmentReplay, "bootstrap_authority", "bootstrap authority has been consumed")
	}
	provided := sha256.Sum256([]byte(authority))
	if !s.bootstrapSet || len(authority) < 32 || subtle.ConstantTimeCompare(provided[:], s.bootstrapDigest[:]) != 1 {
		return ownerRefusal(domain.OwnerEnrollmentInvalid, "bootstrap_authority", "bootstrap authority is missing or invalid")
	}
	return nil
}

func (s *OwnerEnrollmentService) deriveChallenge(instanceID, ceremonyID string) string {
	mac := hmac.New(sha256.New, s.bootstrapDigest[:])
	_, _ = io.WriteString(mac, "nomen-owner-enrollment-v1\x00"+instanceID+"\x00"+ceremonyID)
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *OwnerEnrollmentService) randomID(size int) (string, error) {
	value := make([]byte, size)
	if _, err := io.ReadFull(s.random, value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func (s *OwnerEnrollmentService) recoveryArtifact() (string, error) {
	value, err := s.randomID(32)
	if err != nil {
		return "", err
	}
	return recoveryArtifactPrefix + value, nil
}

type ownerWebAuthnUser struct {
	id, name, displayName string
}

func (u ownerWebAuthnUser) WebAuthnID() []byte                           { return []byte(u.id) }
func (u ownerWebAuthnUser) WebAuthnName() string                         { return u.name }
func (u ownerWebAuthnUser) WebAuthnDisplayName() string                  { return u.displayName }
func (u ownerWebAuthnUser) WebAuthnIcon() string                         { return "" }
func (u ownerWebAuthnUser) WebAuthnCredentials() []gowebauthn.Credential { return nil }

func ownerWebAuthnUserFromEnrollment(enrollment domain.OwnerEnrollment) ownerWebAuthnUser {
	return ownerWebAuthnUser{id: enrollment.OwnerID, name: enrollment.OwnerUsername, displayName: enrollment.OwnerDisplayName}
}

func ownerCreationOptions(issuer string, enrollment domain.OwnerEnrollment, challenge string) (protocol.PublicKeyCredentialCreationOptions, error) {
	server, _, err := ownerWebAuthnServer(issuer)
	if err != nil {
		return protocol.PublicKeyCredentialCreationOptions{}, err
	}
	creation, _, err := server.BeginRegistration(
		ownerWebAuthnUserFromEnrollment(enrollment),
		gowebauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			ResidentKey:      protocol.ResidentKeyRequirementPreferred,
			UserVerification: protocol.VerificationRequired,
		}),
		gowebauthn.WithConveyancePreference(protocol.PreferNoAttestation),
	)
	if err != nil {
		return protocol.PublicKeyCredentialCreationOptions{}, fmt.Errorf("create WebAuthn registration options: %w", err)
	}
	creation.Response.Challenge = protocol.URLEncodedBase64(mustDecodeBase64URL(challenge))
	creation.Response.Timeout = int(ownerCeremonyLifetime / time.Millisecond)
	return creation.Response, nil
}

func ownerWebAuthnServer(issuer string) (*gowebauthn.WebAuthn, string, error) {
	parsed, err := url.Parse(issuer)
	if err != nil || parsed.Hostname() == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Scheme != "https" && !(parsed.Scheme == "http" && parsed.Hostname() == "localhost")) {
		return nil, "", ownerRefusal(domain.OwnerEnrollmentInvalid, "issuer", "issuer is not an allowed WebAuthn origin")
	}
	origin := parsed.Scheme + "://" + parsed.Host
	server, err := gowebauthn.New(&gowebauthn.Config{
		RPID: parsed.Hostname(), RPDisplayName: "Nomen by AngryVibes LLC",
		RPOrigins: []string{origin}, AttestationPreference: protocol.PreferNoAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{ResidentKey: protocol.ResidentKeyRequirementPreferred, UserVerification: protocol.VerificationRequired},
	})
	if err != nil {
		return nil, "", fmt.Errorf("configure WebAuthn relying party: %w", err)
	}
	return server, origin, nil
}

func ownerCredentialFromWebAuthn(credential *gowebauthn.Credential) domain.OwnerCredential {
	transports := make([]string, len(credential.Transport))
	for index, transport := range credential.Transport {
		transports[index] = string(transport)
	}
	var flags byte
	if credential.Flags.UserPresent {
		flags |= 1
	}
	if credential.Flags.UserVerified {
		flags |= 2
	}
	if credential.Flags.BackupEligible {
		flags |= 4
	}
	if credential.Flags.BackupState {
		flags |= 8
	}
	return domain.OwnerCredential{
		ID: credential.ID, PublicKey: credential.PublicKey,
		SignCount: credential.Authenticator.SignCount, AAGUID: credential.Authenticator.AAGUID,
		AttestationType: credential.AttestationType, Transports: transports, Flags: flags,
	}
}

func normalizeOwnerRequest(request BeginOwnerEnrollmentRequest) BeginOwnerEnrollmentRequest {
	request.OwnerID = strings.TrimSpace(request.OwnerID)
	request.Username = strings.TrimSpace(request.Username)
	request.DisplayName = strings.TrimSpace(request.DisplayName)
	return request
}

func validateBeginRequest(request BeginOwnerEnrollmentRequest, idempotencyKey string) error {
	if request.OwnerID == "" || request.Username == "" || request.DisplayName == "" || !utf8.ValidString(request.OwnerID+request.Username+request.DisplayName) {
		return ownerRefusal(domain.OwnerEnrollmentInvalid, "owner", "owner identity is incomplete")
	}
	if len(request.OwnerID) > 200 || len(request.Username) > 320 || len(request.DisplayName) > 200 {
		return ownerRefusal(domain.OwnerEnrollmentInvalid, "owner", "owner identity exceeds its bound")
	}
	if len(idempotencyKey) < 16 || len(idempotencyKey) > 200 {
		return ownerRefusal(domain.OwnerEnrollmentInvalid, "idempotency_key", "idempotency key must be 16 through 200 characters")
	}
	for _, character := range idempotencyKey {
		if character < 0x21 || character > 0x7e {
			return ownerRefusal(domain.OwnerEnrollmentInvalid, "idempotency_key", "idempotency key must use visible ASCII")
		}
	}
	return nil
}

func digestJSON(value any) string {
	encoded, _ := json.Marshal(value)
	return digestBytes(encoded)
}

func digestString(value string) string { return digestBytes([]byte(value)) }
func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func constantDigestEqual(expected, actual string) bool {
	return len(expected) == len(actual) && subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}

func mustDecodeBase64URL(value string) []byte {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		panic("derived challenge is not base64url")
	}
	return decoded
}

func ownerRefusal(reason domain.OwnerEnrollmentRefusal, field, detail string) error {
	return &domain.OwnerEnrollmentError{Reason: reason, Field: field, Detail: detail}
}
