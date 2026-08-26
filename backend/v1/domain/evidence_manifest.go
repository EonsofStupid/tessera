package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const EvidenceManifestSchema = "nomen.evidence-manifest.v1"

type EvidenceLayer string

const (
	EvidenceProductContract       EvidenceLayer = "product_contract"
	EvidenceThreatModel           EvidenceLayer = "threat_model"
	EvidenceSourceLegal           EvidenceLayer = "source_legal"
	EvidencePersistence           EvidenceLayer = "persistence"
	EvidenceServiceImplementation EvidenceLayer = "service_implementation"
	EvidenceAuthorization         EvidenceLayer = "authorization"
	EvidenceProtocolBehavior      EvidenceLayer = "protocol_behavior"
	EvidenceOperatorAPI           EvidenceLayer = "operator_api"
	EvidenceGuidedUI              EvidenceLayer = "guided_ui"
	EvidenceActionEquivalence     EvidenceLayer = "action_equivalence"
	EvidenceAudit                 EvidenceLayer = "audit"
	EvidenceBrowser               EvidenceLayer = "browser"
	EvidenceFailure               EvidenceLayer = "failure"
	EvidenceOperations            EvidenceLayer = "operations"
	EvidenceAccessibility         EvidenceLayer = "accessibility"
	EvidenceDocumentation         EvidenceLayer = "documentation"
	EvidenceSupplyChain           EvidenceLayer = "supply_chain"
	EvidenceReleasePolicy         EvidenceLayer = "release_policy"
)

var (
	requiredEvidenceLayers = []EvidenceLayer{
		EvidenceProductContract, EvidenceThreatModel, EvidenceSourceLegal,
		EvidencePersistence, EvidenceServiceImplementation, EvidenceAuthorization,
		EvidenceProtocolBehavior, EvidenceOperatorAPI, EvidenceGuidedUI,
		EvidenceActionEquivalence, EvidenceAudit, EvidenceBrowser, EvidenceFailure,
		EvidenceOperations, EvidenceAccessibility, EvidenceDocumentation,
		EvidenceSupplyChain, EvidenceReleasePolicy,
	}
	sourceRevisionPattern = regexp.MustCompile(`^[0-9a-f]{40,64}$`)
	conformanceIDPattern  = regexp.MustCompile(`^nomen[.]conformance[.][a-z][a-z0-9_.-]*[.]v[0-9]+$`)
)

type DeploymentProfile string

const (
	DeploymentStandalone     DeploymentProfile = "standalone"
	DeploymentShippinManaged DeploymentProfile = "shippin_managed"
)

func (p DeploymentProfile) Valid() bool {
	return p == DeploymentStandalone || p == DeploymentShippinManaged
}

type EvidenceManifest struct {
	Schema                   string               `json:"schema"`
	ConformanceID            string               `json:"conformance_id"`
	CapabilityID             string               `json:"capability_id"`
	CapabilityLedgerRevision string               `json:"capability_ledger_revision"`
	ReleaseID                string               `json:"release_id"`
	SourceRevision           string               `json:"source_revision"`
	ImageDigest              string               `json:"image_digest"`
	BundleManifestDigest     string               `json:"bundle_manifest_digest"`
	DeploymentProfile        DeploymentProfile    `json:"deployment_profile"`
	EvidenceOwner            string               `json:"evidence_owner"`
	Result                   ConformanceResult    `json:"result"`
	VerifiedAt               time.Time            `json:"verified_at"`
	ValidUntil               *time.Time           `json:"valid_until,omitempty"`
	Evidence                 []EvidenceAssertion  `json:"evidence"`
	Limitations              []EvidenceLimitation `json:"limitations,omitempty"`
	ExternalAssurances       []ExternalAssurance  `json:"external_assurances,omitempty"`
}

type EvidenceAssertion struct {
	Layer              EvidenceLayer     `json:"layer"`
	Result             ConformanceResult `json:"result"`
	ArtifactDigest     string            `json:"artifact_digest"`
	Observations       uint64            `json:"observations"`
	ProtectedReference string            `json:"protected_reference,omitempty"`
}

type EvidenceLimitation struct {
	ID                  string `json:"id"`
	Impact              string `json:"impact"`
	OperatorRemediation string `json:"operator_remediation"`
}

type ExternalAssurance struct {
	Track      string     `json:"track"`
	Issuer     string     `json:"issuer"`
	Identifier string     `json:"identifier"`
	Scope      string     `json:"scope"`
	IssuedAt   time.Time  `json:"issued_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

func RequiredEvidenceLayers() []EvidenceLayer {
	return append([]EvidenceLayer(nil), requiredEvidenceLayers...)
}

func (m EvidenceManifest) Validate() error {
	if m.Schema != EvidenceManifestSchema || !conformanceIDPattern.MatchString(m.ConformanceID) {
		return fmt.Errorf("evidence manifest identity is incomplete")
	}
	ledger, err := LoadCapabilityLedger()
	if err != nil {
		return err
	}
	entry, found := capabilityLedgerEntry(ledger, m.CapabilityID)
	if !found {
		return fmt.Errorf("evidence manifest references unknown capability %q", m.CapabilityID)
	}
	if m.CapabilityLedgerRevision != ledger.ProgramRevision {
		return fmt.Errorf("capability ledger revision does not match embedded policy")
	}
	if strings.TrimSpace(m.ReleaseID) == "" || !sourceRevisionPattern.MatchString(m.SourceRevision) {
		return fmt.Errorf("release_id and lowercase source_revision are required")
	}
	if !validPlanDigest(m.ImageDigest) || !validPlanDigest(m.BundleManifestDigest) {
		return fmt.Errorf("image and bundle manifest digests must be lowercase sha256 digests")
	}
	if !m.DeploymentProfile.Valid() || m.EvidenceOwner != entry.EvidenceOwner || !m.Result.Valid() || m.VerifiedAt.IsZero() {
		return fmt.Errorf("evidence manifest release policy is incomplete")
	}
	if m.ValidUntil != nil && !m.ValidUntil.After(m.VerifiedAt) {
		return fmt.Errorf("valid_until must be after verified_at")
	}
	if err := validateEvidenceAssertions(m.Evidence, m.Result); err != nil {
		return err
	}
	if err := validateEvidenceLimitations(m.Limitations); err != nil {
		return err
	}
	return validateExternalAssurances(m.ExternalAssurances, m.VerifiedAt)
}

func (m EvidenceManifest) EvidenceDigest() (string, error) {
	if err := m.Validate(); err != nil {
		return "", err
	}
	canonical, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("encode evidence manifest: %w", err)
	}
	digest := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func (m EvidenceManifest) CapabilityProof() (*CapabilityProof, error) {
	if m.Result != ConformancePassed {
		return nil, fmt.Errorf("failed evidence manifest cannot produce capability proof")
	}
	digest, err := m.EvidenceDigest()
	if err != nil {
		return nil, err
	}
	return &CapabilityProof{
		ConformanceID: m.ConformanceID, BundleManifestDigest: m.BundleManifestDigest,
		Result: m.Result, VerifiedAt: m.VerifiedAt, EvidenceDigest: digest,
	}, nil
}

func capabilityLedgerEntry(ledger CapabilityLedger, id string) (CapabilityLedgerEntry, bool) {
	for _, entry := range ledger.Entries() {
		if entry.ID == id {
			return entry, true
		}
	}
	return CapabilityLedgerEntry{}, false
}

func validateEvidenceAssertions(assertions []EvidenceAssertion, manifestResult ConformanceResult) error {
	required := make(map[EvidenceLayer]struct{}, len(requiredEvidenceLayers))
	for _, layer := range requiredEvidenceLayers {
		required[layer] = struct{}{}
	}
	seen := make(map[EvidenceLayer]struct{}, len(assertions))
	for _, assertion := range assertions {
		if _, known := required[assertion.Layer]; !known {
			return fmt.Errorf("unknown evidence layer %q", assertion.Layer)
		}
		if _, duplicate := seen[assertion.Layer]; duplicate {
			return fmt.Errorf("duplicate evidence layer %q", assertion.Layer)
		}
		seen[assertion.Layer] = struct{}{}
		if !assertion.Result.Valid() || !validPlanDigest(assertion.ArtifactDigest) || assertion.Observations == 0 {
			return fmt.Errorf("evidence layer %s is incomplete", assertion.Layer)
		}
		if manifestResult == ConformancePassed && assertion.Result != ConformancePassed {
			return fmt.Errorf("passing manifest contains failed evidence layer %s", assertion.Layer)
		}
	}
	for _, layer := range requiredEvidenceLayers {
		if _, present := seen[layer]; !present {
			return fmt.Errorf("required evidence layer %s is missing", layer)
		}
	}
	return nil
}

func validateEvidenceLimitations(limitations []EvidenceLimitation) error {
	seen := make(map[string]struct{}, len(limitations))
	for _, limitation := range limitations {
		if !capabilityIDPattern.MatchString(limitation.ID) || strings.TrimSpace(limitation.Impact) == "" || strings.TrimSpace(limitation.OperatorRemediation) == "" {
			return fmt.Errorf("evidence limitation is incomplete")
		}
		if _, duplicate := seen[limitation.ID]; duplicate {
			return fmt.Errorf("duplicate evidence limitation %q", limitation.ID)
		}
		seen[limitation.ID] = struct{}{}
	}
	return nil
}

func validateExternalAssurances(assurances []ExternalAssurance, verifiedAt time.Time) error {
	seen := make(map[string]struct{}, len(assurances))
	for _, assurance := range assurances {
		if !capabilityIDPattern.MatchString(assurance.Track) || strings.TrimSpace(assurance.Issuer) == "" || strings.TrimSpace(assurance.Identifier) == "" || strings.TrimSpace(assurance.Scope) == "" || assurance.IssuedAt.IsZero() {
			return fmt.Errorf("external assurance is incomplete")
		}
		if assurance.IssuedAt.After(verifiedAt) {
			return fmt.Errorf("external assurance %s was issued after manifest verification", assurance.Track)
		}
		if assurance.ExpiresAt != nil && !assurance.ExpiresAt.After(assurance.IssuedAt) {
			return fmt.Errorf("external assurance %s has invalid expiry", assurance.Track)
		}
		key := assurance.Track + "\x00" + assurance.Identifier
		if _, duplicate := seen[key]; duplicate {
			return fmt.Errorf("duplicate external assurance %s", assurance.Track)
		}
		seen[key] = struct{}{}
	}
	return nil
}
