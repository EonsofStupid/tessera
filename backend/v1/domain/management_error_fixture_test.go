package domain

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type managementRemedyFixture struct {
	Type           ManagementErrorType `json:"type"`
	HTTPStatus     int                 `json:"http_status"`
	GRPCCode       string              `json:"grpc_code"`
	Title          string              `json:"title"`
	Consequence    string              `json:"consequence"`
	Action         ManagementRemedy    `json:"action"`
	RequiredDetail string              `json:"required_detail"`
}

func TestManagementRemedyFixturesCoverCatalog(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("../../../testdata/tessera/management-error-remedies.json")
	require.NoError(t, err)

	var fixtures []managementRemedyFixture
	require.NoError(t, json.Unmarshal(contents, &fixtures))
	require.Len(t, fixtures, len(ManagementErrorCatalog()))

	titles := make(map[string]struct{}, len(fixtures))
	actions := make(map[ManagementRemedyKind]struct{}, len(fixtures))
	seen := make(map[ManagementErrorType]struct{}, len(fixtures))
	requiredDetails := map[ManagementErrorType]string{
		ManagementErrorEntitlementRequired: "missing_entitlement",
		ManagementErrorPermissionRequired:  "required_permission",
		ManagementErrorStepUpRequired:      "required_assurance",
		ManagementErrorRateLimited:         "retry_after_seconds",
		ManagementErrorServiceUnavailable:  "diagnostic_ref",
	}
	for _, fixture := range fixtures {
		spec, ok := ManagementErrorSpecFor(fixture.Type)
		require.True(t, ok, "unknown fixture type %q", fixture.Type)
		_, duplicate := seen[fixture.Type]
		assert.False(t, duplicate, "duplicate fixture type %q", fixture.Type)
		seen[fixture.Type] = struct{}{}
		assert.Equal(t, spec.HTTPStatus, fixture.HTTPStatus)
		assert.Equal(t, spec.GRPCCode, fixture.GRPCCode)
		assert.Equal(t, spec.Remedy, fixture.Action.Kind)
		assert.NotEmpty(t, fixture.Action.Label)
		assert.NotEmpty(t, fixture.Title)
		assert.NotEmpty(t, fixture.Consequence)
		assert.Equal(t, requiredDetails[fixture.Type], fixture.RequiredDetail)
		titles[fixture.Title] = struct{}{}
		actions[fixture.Action.Kind] = struct{}{}
	}

	assert.Len(t, titles, len(fixtures), "every error type needs a distinct title")
	assert.Len(t, actions, len(fixtures), "every error type needs a distinct primary action")
	assert.Equal(t, 403, fixtureFor(t, fixtures, ManagementErrorEntitlementRequired).HTTPStatus)
}

func fixtureFor(t *testing.T, fixtures []managementRemedyFixture, errorType ManagementErrorType) managementRemedyFixture {
	t.Helper()
	for _, fixture := range fixtures {
		if fixture.Type == errorType {
			return fixture
		}
	}
	require.FailNow(t, "missing fixture", "type %q", errorType)
	return managementRemedyFixture{}
}
