package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOperatorActionCatalogRejectsExecutableMarkup(t *testing.T) {
	t.Parallel()
	catalog := OperatorActionCatalog{
		SchemaVersion: 1, ResourceRevision: "sha256:test", ObservedAt: time.Now(),
		Actions: []OperatorAction{{
			ID: "action.provider_plan", Title: "Plan provider", Consequence: "Creates a reviewed provider plan.",
			Stage: OperatorActionPlan, Method: "POST", Href: "/nomen/v1/providers:plan",
			IntentSchema: json.RawMessage(`{"type":"object"}`), RequiredPermissions: []string{"nomen.providers.plan"},
			CapabilityID: CapabilityIDUpstreamOIDC, Exposure: UIExposureDisabled, Reason: "conformance_pending",
		}},
	}
	require.NoError(t, catalog.Validate())
	catalog.Actions[0].IntentSchema = json.RawMessage(`<script>unsafe()</script>`)
	require.Error(t, catalog.Validate())
}
