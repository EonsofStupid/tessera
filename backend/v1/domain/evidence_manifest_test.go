package domain

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPassingEvidenceManifestProducesBundleBoundProof(t *testing.T) {
	manifest := passingEvidenceManifest(t)
	require.NoError(t, manifest.Validate())

	proof, err := manifest.CapabilityProof()
	require.NoError(t, err)
	assert.Equal(t, manifest.ConformanceID, proof.ConformanceID)
	assert.Equal(t, manifest.BundleManifestDigest, proof.BundleManifestDigest)
	assert.Equal(t, ConformancePassed, proof.Result)
	assert.Regexp(t, `^sha256:[0-9a-f]{64}$`, proof.EvidenceDigest)

	discovery := validCapabilityDiscovery()
	discovery.BundleManifestDigest = manifest.BundleManifestDigest
	discovery.Capabilities[0].Proof = proof
	require.NoError(t, discovery.Validate())
}

func TestEvidenceManifestFailsClosed(t *testing.T) {
	tests := []struct {
		name string
		edit func(*EvidenceManifest)
		want string
	}{
		{"unknown capability", func(m *EvidenceManifest) { m.CapabilityID = "imaginary" }, "unknown capability"},
		{"ledger drift", func(m *EvidenceManifest) { m.CapabilityLedgerRevision = "old" }, "ledger revision"},
		{"image digest", func(m *EvidenceManifest) { m.ImageDigest = "latest" }, "lowercase sha256"},
		{"owner mismatch", func(m *EvidenceManifest) { m.EvidenceOwner = "someone_else" }, "release policy"},
		{"missing layer", func(m *EvidenceManifest) { m.Evidence = m.Evidence[1:] }, "required evidence layer"},
		{"failed layer", func(m *EvidenceManifest) { m.Evidence[0].Result = ConformanceFailed }, "passing manifest contains failed"},
		{"empty observation", func(m *EvidenceManifest) { m.Evidence[0].Observations = 0 }, "is incomplete"},
		{"expired before verification", func(m *EvidenceManifest) { value := m.VerifiedAt.Add(-time.Hour); m.ValidUntil = &value }, "valid_until"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := passingEvidenceManifest(t)
			test.edit(&manifest)
			err := manifest.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.want)
		})
	}
}

func TestFailedEvidenceManifestCannotProduceProof(t *testing.T) {
	manifest := passingEvidenceManifest(t)
	manifest.Result = ConformanceFailed
	manifest.Evidence[0].Result = ConformanceFailed
	require.NoError(t, manifest.Validate())
	_, err := manifest.CapabilityProof()
	require.ErrorContains(t, err, "cannot produce capability proof")
}

func TestEvidenceManifestSchemaDocumentIsValidJSON(t *testing.T) {
	contents, err := os.ReadFile("evidence-manifest.schema.json")
	require.NoError(t, err)
	var schema map[string]any
	require.NoError(t, json.Unmarshal(contents, &schema))
	assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", schema["$schema"])
}

func passingEvidenceManifest(t *testing.T) EvidenceManifest {
	t.Helper()
	ledger, err := LoadCapabilityLedger()
	require.NoError(t, err)
	verifiedAt := time.Date(2026, time.August, 21, 20, 0, 0, 0, time.UTC)
	digest := "sha256:" + string(make([]byte, 0)) + "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	evidence := make([]EvidenceAssertion, 0, len(RequiredEvidenceLayers()))
	for _, layer := range RequiredEvidenceLayers() {
		evidence = append(evidence, EvidenceAssertion{Layer: layer, Result: ConformancePassed, ArtifactDigest: digest, Observations: 1})
	}
	return EvidenceManifest{
		Schema: EvidenceManifestSchema, ConformanceID: "nomen.conformance.downstream_oidc.v1",
		CapabilityID: CapabilityIDDownstreamOIDC, CapabilityLedgerRevision: ledger.ProgramRevision,
		ReleaseID: "2026.8.0-rc.1", SourceRevision: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ImageDigest: digest, BundleManifestDigest: digest, DeploymentProfile: DeploymentStandalone,
		EvidenceOwner: "identity_protocols", Result: ConformancePassed, VerifiedAt: verifiedAt, Evidence: evidence,
	}
}
