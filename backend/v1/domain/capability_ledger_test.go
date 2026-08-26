package domain

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedCapabilityLedgerIsValidAndComplete(t *testing.T) {
	ledger, err := LoadCapabilityLedger()
	require.NoError(t, err)
	require.NoError(t, ledger.Validate())

	union := make(map[string]struct{})
	supporting := make(map[string]struct{})
	for _, entry := range ledger.Entries() {
		assert.Equal(t, CapabilityPreview, entry.CurrentStatus, entry.ID)
		assert.Contains(t, entry.RequiredComponents, ComponentNomen, entry.ID)
		switch entry.ParityScope {
		case "union":
			union[entry.ID] = struct{}{}
		case "supporting":
			supporting[entry.ID] = struct{}{}
		default:
			t.Fatalf("unknown parity scope %q", entry.ParityScope)
		}
	}

	documented := documentedUnionCapabilityIDs(t)
	assert.Equal(t, documented, union, "the active parity contract and machine ledger drifted")
	assert.Equal(t, map[string]struct{}{
		CapabilityIDOverview: {}, CapabilityIDAnalyticsOLAP: {}, CapabilityIDVaultixSecretCustody: {},
	}, supporting)
}

func TestCapabilityLedgerSchemaDocumentIsValidJSON(t *testing.T) {
	contents, err := os.ReadFile("capability-ledger.schema.json")
	require.NoError(t, err)
	var schema map[string]any
	require.NoError(t, json.Unmarshal(contents, &schema))
	assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", schema["$schema"])
}

func TestLoadCapabilityLedgerReturnsDetachedData(t *testing.T) {
	first, err := LoadCapabilityLedger()
	require.NoError(t, err)
	first.Families[0].Capabilities[0].ID = "mutated"
	first.Families[0].RequiredComponents[0] = ComponentZuul

	second, err := LoadCapabilityLedger()
	require.NoError(t, err)
	assert.NotEqual(t, "mutated", second.Families[0].Capabilities[0].ID)
	assert.Equal(t, ComponentNomen, second.Families[0].RequiredComponents[0])
}

func documentedUnionCapabilityIDs(t *testing.T) map[string]struct{} {
	t.Helper()
	contents, err := os.ReadFile("../../../docs/22-certification-and-parity-program.md")
	require.NoError(t, err)
	contract := string(contents)
	start := strings.Index(contract, "## Union capability ledger")
	end := strings.Index(contract[start:], "## Delivery sequence")
	require.NotEqual(t, -1, start)
	require.NotEqual(t, -1, end)
	section := contract[start : start+end]

	matches := regexp.MustCompile("`([a-z][a-z0-9_]+)`").FindAllStringSubmatch(section, -1)
	ids := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		ids[match[1]] = struct{}{}
	}
	return ids
}
