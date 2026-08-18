package domain

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestManagementErrorWireContractMatchesDomain(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("../../../proto/tessera/management/v1/error.proto")
	require.NoError(t, err)
	wire := string(contents)

	for _, spec := range ManagementErrorCatalog() {
		require.Contains(t, wire, "MANAGEMENT_ERROR_TYPE_"+strings.ToUpper(string(spec.Type)))
		require.Contains(t, wire, "MANAGEMENT_REMEDY_KIND_"+strings.ToUpper(string(spec.Remedy)))
	}
	for _, field := range []string{
		"missing_entitlement",
		"required_permission",
		"required_assurance",
		"current_revision",
		"retry_after_seconds",
		"diagnostic_ref",
	} {
		require.Contains(t, wire, field)
	}
}
