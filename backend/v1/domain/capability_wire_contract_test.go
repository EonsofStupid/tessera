package domain

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCapabilityWireContractMatchesDomain(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("../../../proto/nomen/management/v1/capability.proto")
	require.NoError(t, err)
	wire := string(contents)

	for _, value := range []ComponentRole{ComponentNomen, ComponentNomenOperator, ComponentClickHouse, ComponentVaultix, ComponentZuul, ComponentShippinAdapter} {
		require.Contains(t, wire, "COMPONENT_ROLE_"+strings.ToUpper(string(value)))
	}
	for _, value := range []CompatibilityState{CompatibilityCompatible, CompatibilityIncompatible, CompatibilityNotPresent, CompatibilityUnknown} {
		require.Contains(t, wire, "COMPATIBILITY_STATE_"+strings.ToUpper(string(value)))
	}
	for _, value := range []CapabilityStatus{CapabilityUnsupported, CapabilityPreview, CapabilityOperational, CapabilityDegraded} {
		require.Contains(t, wire, "CAPABILITY_STATUS_"+strings.ToUpper(string(value)))
	}
	for _, value := range []UIExposure{UIExposureHidden, UIExposureDisabled, UIExposureEnabled} {
		require.Contains(t, wire, "UI_EXPOSURE_"+strings.ToUpper(string(value)))
	}
	for _, value := range []ConformanceResult{ConformancePassed, ConformanceFailed} {
		require.Contains(t, wire, "CONFORMANCE_RESULT_"+strings.ToUpper(string(value)))
	}
	for _, field := range []string{
		"schema_version",
		"resource_revision",
		"bundle_manifest_digest",
		"required_components",
		"operation_kinds",
		"conformance_id",
		"verified_at",
		"evidence_digest",
	} {
		require.Contains(t, wire, field)
	}
}
