package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var capabilityTestNow = time.Date(2026, time.August, 18, 12, 0, 0, 0, time.UTC)

func TestCapabilityDiscoveryValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(*CapabilityDiscovery)
		wantErr string
	}{
		{"valid", func(*CapabilityDiscovery) {}, ""},
		{"schema", func(value *CapabilityDiscovery) { value.SchemaVersion = 2 }, "unsupported discovery schema"},
		{"revision", func(value *CapabilityDiscovery) { value.ResourceRevision = "" }, "resource_revision"},
		{"observation", func(value *CapabilityDiscovery) { value.ObservedAt = time.Time{} }, "observed_at"},
		{"digest", func(value *CapabilityDiscovery) { value.BundleManifestDigest = "sha256:no" }, "bundle_manifest_digest"},
		{"component role", func(value *CapabilityDiscovery) { value.Components[0].Role = "future" }, "unknown component role"},
		{"duplicate component", func(value *CapabilityDiscovery) { value.Components = append(value.Components, value.Components[0]) }, "duplicate component"},
		{"component observation", func(value *CapabilityDiscovery) { value.Components[0].ObservedAt = time.Time{} }, "incomplete compatibility"},
		{"component version", func(value *CapabilityDiscovery) { value.Components[0].Version = "" }, "incomplete version"},
		{"component state", func(value *CapabilityDiscovery) { value.Components[0].State = "future" }, "incomplete compatibility"},
		{"component reason", func(value *CapabilityDiscovery) { value.Components[0].State = CompatibilityUnknown }, "compatibility reason"},
		{"missing mandatory", func(value *CapabilityDiscovery) { value.Components = value.Components[1:] }, "mandatory component tessera"},
		{"capability identity", func(value *CapabilityDiscovery) { value.Capabilities[0].ID = "" }, "incomplete identity"},
		{"duplicate capability", func(value *CapabilityDiscovery) {
			value.Capabilities = append(value.Capabilities, value.Capabilities[0])
		}, "duplicate capability"},
		{"disabled reason", func(value *CapabilityDiscovery) { value.Capabilities[1].Reason = "" }, "requires a reason"},
		{"unsupported enabled", func(value *CapabilityDiscovery) { value.Capabilities[1].Exposure = UIExposureEnabled }, "cannot enable status"},
		{"available without proof", func(value *CapabilityDiscovery) { value.Capabilities[0].Proof = nil }, "requires passing conformance proof"},
		{"failed proof", func(value *CapabilityDiscovery) { value.Capabilities[0].Proof.Result = ConformanceFailed }, "requires passing conformance proof"},
		{"proof identity", func(value *CapabilityDiscovery) { value.Capabilities[0].Proof.ConformanceID = "" }, "incomplete conformance proof"},
		{"proof bundle", func(value *CapabilityDiscovery) {
			value.Capabilities[0].Proof.BundleManifestDigest = "sha256:" + strings.Repeat("b", 64)
		}, "does not match installed bundle"},
		{"proof evidence", func(value *CapabilityDiscovery) { value.Capabilities[0].Proof.EvidenceDigest = "sha256:no" }, "proof digests"},
		{"component required", func(value *CapabilityDiscovery) {
			value.Capabilities[0].RequiredComponents = append(value.Capabilities[0].RequiredComponents, "future")
		}, "unknown required component"},
		{"component repeated", func(value *CapabilityDiscovery) {
			value.Capabilities[0].RequiredComponents = append(value.Capabilities[0].RequiredComponents, ComponentTessera)
		}, "repeats required component"},
		{"operation kind", func(value *CapabilityDiscovery) {
			value.Capabilities[0].OperationKinds = append(value.Capabilities[0].OperationKinds, "future")
		}, "unknown operation kind"},
		{"operation repeated", func(value *CapabilityDiscovery) {
			value.Capabilities[0].OperationKinds = append(value.Capabilities[0].OperationKinds, OperationInstallation)
		}, "repeats operation kind"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := validCapabilityDiscovery()
			test.mutate(&value)
			err := value.Validate()
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func TestCapabilityDiscoveryAllowsAbsentOptionalZuul(t *testing.T) {
	t.Parallel()

	value := validCapabilityDiscovery()
	value.Components[2] = ComponentCompatibility{
		Role:       ComponentZuul,
		State:      CompatibilityNotPresent,
		Reason:     "not installed",
		ObservedAt: capabilityTestNow,
	}
	require.NoError(t, value.Validate())
}

func TestCapabilityDiscoveryAllowsAbsentOptionalVaultix(t *testing.T) {
	t.Parallel()

	value := validCapabilityDiscovery()
	value.Components[3] = ComponentCompatibility{
		Role:       ComponentVaultix,
		State:      CompatibilityNotPresent,
		Reason:     "not installed",
		ObservedAt: capabilityTestNow,
	}
	require.NoError(t, value.Validate())
}

func TestResolveCapability(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mutate     func(*CapabilityDiscovery)
		schemas    []uint32
		capability string
		want       CapabilityResolution
	}{
		{"enabled", func(*CapabilityDiscovery) {}, []uint32{1}, "installation", CapabilityResolution{"installation", UIExposureEnabled, ""}},
		{"schema incompatible", func(value *CapabilityDiscovery) { value.SchemaVersion = 2 }, []uint32{1}, "installation", CapabilityResolution{"installation", UIExposureDisabled, "schema_incompatible"}},
		{"invalid discovery", func(value *CapabilityDiscovery) { value.ResourceRevision = "" }, []uint32{1}, "installation", CapabilityResolution{"installation", UIExposureDisabled, "invalid_discovery"}},
		{"not advertised", func(*CapabilityDiscovery) {}, []uint32{1}, "future", CapabilityResolution{"future", UIExposureHidden, "not_advertised"}},
		{"server hidden", func(*CapabilityDiscovery) {}, []uint32{1}, "backup", CapabilityResolution{"backup", UIExposureHidden, "not_configured"}},
		{"required component incompatible", func(value *CapabilityDiscovery) {
			value.Components[2].State = CompatibilityIncompatible
			value.Components[2].Reason = "bundle mismatch"
		}, []uint32{1}, "zuul_enrollment", CapabilityResolution{"zuul_enrollment", UIExposureDisabled, "component_unavailable"}},
		{"required component absent", func(value *CapabilityDiscovery) { value.Components = value.Components[:2] }, []uint32{1}, "zuul_enrollment", CapabilityResolution{"zuul_enrollment", UIExposureDisabled, "component_unavailable"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := validCapabilityDiscovery()
			test.mutate(&value)
			assert.Equal(t, test.want, ResolveCapability(value, test.schemas, test.capability))
		})
	}
}

func TestCapabilityVocabulary(t *testing.T) {
	t.Parallel()

	for _, value := range []ComponentRole{ComponentTessera, ComponentShippinAdapter, ComponentZuul, ComponentVaultix} {
		assert.True(t, value.Valid())
	}
	assert.False(t, ComponentRole("future").Valid())
	for _, value := range []CompatibilityState{CompatibilityCompatible, CompatibilityIncompatible, CompatibilityNotPresent, CompatibilityUnknown} {
		assert.True(t, value.Valid())
	}
	assert.False(t, CompatibilityState("future").Valid())
	for _, value := range []CapabilityStatus{CapabilityUnsupported, CapabilityPreview, CapabilityAvailable, CapabilityDegraded} {
		assert.True(t, value.Valid())
	}
	assert.False(t, CapabilityStatus("future").Valid())
	for _, value := range []UIExposure{UIExposureHidden, UIExposureDisabled, UIExposureEnabled} {
		assert.True(t, value.Valid())
	}
	assert.False(t, UIExposure("future").Valid())
	for _, value := range []ConformanceResult{ConformancePassed, ConformanceFailed} {
		assert.True(t, value.Valid())
	}
	assert.False(t, ConformanceResult("future").Valid())
	assert.Equal(t, "ldap_outbound", CapabilityIDLDAPOutbound)
	assert.Equal(t, "ldap_inbound", CapabilityIDLDAPInbound)
	assert.Equal(t, "forward_auth", CapabilityIDForwardAuth)
	assert.Equal(t, "identity_aware_proxy", CapabilityIDIdentityAwareProxy)
	assert.Equal(t, "visual_flow_engine", CapabilityIDVisualFlowEngine)
	assert.Equal(t, "vaultix_secret_custody", CapabilityIDVaultixSecretCustody)
}

func validCapabilityDiscovery() CapabilityDiscovery {
	bundleDigest := "sha256:" + strings.Repeat("a", 64)
	return CapabilityDiscovery{
		SchemaVersion:        1,
		ResourceRevision:     "capabilities-01",
		BundleManifestDigest: bundleDigest,
		ObservedAt:           capabilityTestNow,
		Components: []ComponentCompatibility{
			{Role: ComponentTessera, Version: "1.4.0", APIMajor: 1, State: CompatibilityCompatible, ObservedAt: capabilityTestNow},
			{Role: ComponentShippinAdapter, Version: "2.1.0", APIMajor: 1, State: CompatibilityCompatible, ObservedAt: capabilityTestNow},
			{Role: ComponentZuul, Version: "0.8.0", APIMajor: 1, State: CompatibilityCompatible, ObservedAt: capabilityTestNow},
			{Role: ComponentVaultix, Version: "0.1.0", APIMajor: 1, State: CompatibilityCompatible, ObservedAt: capabilityTestNow},
		},
		Capabilities: []CapabilityFact{
			{ID: "installation", Status: CapabilityAvailable, Exposure: UIExposureEnabled, RequiredComponents: []ComponentRole{ComponentTessera, ComponentShippinAdapter}, OperationKinds: []OperationKind{OperationInstallation}, Proof: passingCapabilityProof("tessera.conformance.installation.v1", bundleDigest)},
			{ID: "backup", Status: CapabilityUnsupported, Exposure: UIExposureHidden, Reason: "not_configured", RequiredComponents: []ComponentRole{ComponentTessera}, OperationKinds: []OperationKind{OperationBackup}},
			{ID: "zuul_enrollment", Status: CapabilityAvailable, Exposure: UIExposureEnabled, RequiredComponents: []ComponentRole{ComponentTessera, ComponentShippinAdapter, ComponentZuul}, OperationKinds: []OperationKind{OperationGuide}, Proof: passingCapabilityProof("tessera.conformance.zuul-enrollment.v1", bundleDigest)},
		},
	}
}

func passingCapabilityProof(conformanceID, bundleDigest string) *CapabilityProof {
	return &CapabilityProof{
		ConformanceID:        conformanceID,
		BundleManifestDigest: bundleDigest,
		Result:               ConformancePassed,
		VerifiedAt:           capabilityTestNow,
		EvidenceDigest:       "sha256:" + strings.Repeat("c", 64),
	}
}
