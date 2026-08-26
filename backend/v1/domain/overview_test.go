package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildOverviewReadyAndStable(t *testing.T) {
	facts := OverviewFacts{
		WorkspaceAttachments: 7,
		AgentSeats:           4,
		HumanSeats:           18,
		Flows:                3,
		PolicyRevisions:      []string{"pol-b", "pol-a", "pol-b"},
	}
	first := BuildOverview(facts, "https://id.nomen.test", 1, time.Unix(100, 0))
	second := BuildOverview(facts, "https://id.nomen.test", 1, time.Unix(200, 0))
	reordered := facts
	reordered.PolicyRevisions = []string{"pol-a", "pol-b"}
	third := BuildOverview(reordered, "https://id.nomen.test", 1, time.Unix(300, 0))

	require.NoError(t, first.Validate())
	assert.Equal(t, OverviewReady, first.Readiness.Status)
	assert.Empty(t, first.Readiness.Reasons)
	assert.Equal(t, first.ResourceRevision, second.ResourceRevision, "observation time is not source state")
	assert.Equal(t, first.ResourceRevision, third.ResourceRevision, "policy revisions are a set")
	assert.NotEqual(t, first.ObservedAt, second.ObservedAt)
	assert.NotNil(t, first.Federation.Upstreams)
	assert.NotNil(t, first.Federation.Clients)
	assert.NotNil(t, first.Activity)
}

func TestBuildOverviewNeverPromotesMissingFacts(t *testing.T) {
	overview := BuildOverview(OverviewFacts{}, "http://127.0.0.1:8080", 0, time.Now())

	require.NoError(t, overview.Validate())
	assert.Equal(t, OverviewDegraded, overview.Readiness.Status)
	assert.ElementsMatch(t, []string{
		"no_active_asymmetric_signing_key",
		"no_nomen_flow_configured",
	}, overview.Readiness.Reasons)
	for _, lens := range overview.Lenses {
		assert.Equal(t, OverviewLensQuiet, lens.Status)
	}
}

func TestOverviewRejectsFalseReadyState(t *testing.T) {
	overview := BuildOverview(OverviewFacts{Flows: 1}, "https://id.nomen.test", 1, time.Now())
	overview.Readiness.SigningKeys = 0

	require.ErrorContains(t, overview.Validate(), "ready overview requires")
}
