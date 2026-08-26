package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildDeploymentPreflightEvaluatesCanonicalChecks(t *testing.T) {
	now := time.Date(2026, time.August, 22, 1, 0, 0, 0, time.UTC)
	preflight := BuildDeploymentPreflight(DeploymentPreflightFacts{
		DatabaseAvailable: true, DatabaseProbeAvailable: true,
		SigningKeys: 1, SigningProbeAvailable: true,
		NotificationConfigured: true, NotificationProbeAvailable: true,
	}, "https://identity.example.test", now)

	require.NoError(t, preflight.Validate())
	assert.Equal(t, PreflightReady, preflight.Status)
	assert.Equal(t, preflightCheckOrder, []string{preflight.Checks[0].ID, preflight.Checks[1].ID, preflight.Checks[2].ID, preflight.Checks[3].ID, preflight.Checks[4].ID})
	assert.Equal(t, now, preflight.ObservedAt)
}

func TestBuildDeploymentPreflightAllowsOnlyLoopbackHTTPAsWarning(t *testing.T) {
	facts := DeploymentPreflightFacts{DatabaseAvailable: true, DatabaseProbeAvailable: true, SigningKeys: 1, SigningProbeAvailable: true, NotificationConfigured: true, NotificationProbeAvailable: true}
	loopback := BuildDeploymentPreflight(facts, "http://localhost:8080", time.Now())
	require.NoError(t, loopback.Validate())
	assert.Equal(t, PreflightActionRequired, loopback.Status)
	assert.Equal(t, PreflightCheckWarning, loopback.Checks[2].Status)
	assert.False(t, loopback.Checks[2].Required)

	remote := BuildDeploymentPreflight(facts, "http://identity.example.test", time.Now())
	require.NoError(t, remote.Validate())
	assert.Equal(t, PreflightBlocked, remote.Status)
	assert.Equal(t, PreflightCheckFailed, remote.Checks[2].Status)
}

func TestBuildDeploymentPreflightRejectsIPLiteralWebAuthnIssuer(t *testing.T) {
	facts := DeploymentPreflightFacts{DatabaseAvailable: true, DatabaseProbeAvailable: true, SigningKeys: 1, SigningProbeAvailable: true, NotificationConfigured: true, NotificationProbeAvailable: true}
	preflight := BuildDeploymentPreflight(facts, "https://127.0.0.1:8080", time.Now())
	require.NoError(t, preflight.Validate())
	assert.Equal(t, PreflightBlocked, preflight.Status)
	assert.Equal(t, "webauthn_rp_id_invalid", preflight.Checks[1].Reason)
}

func TestBuildDeploymentPreflightFailsClosedWithoutDatabaseOrSigningEvidence(t *testing.T) {
	preflight := BuildDeploymentPreflight(DeploymentPreflightFacts{}, "https://identity.example.test", time.Now())
	require.NoError(t, preflight.Validate())
	assert.Equal(t, PreflightBlocked, preflight.Status)
	assert.Equal(t, "database_probe_unavailable", preflight.Checks[0].Reason)
	assert.Equal(t, "signing_probe_unavailable", preflight.Checks[3].Reason)
	assert.Equal(t, PreflightCheckWarning, preflight.Checks[4].Status)
}

func TestDeploymentPreflightRevisionIgnoresObservationTime(t *testing.T) {
	facts := DeploymentPreflightFacts{DatabaseAvailable: true, DatabaseProbeAvailable: true, SigningKeys: 1, SigningProbeAvailable: true, NotificationConfigured: false, NotificationProbeAvailable: true}
	first := BuildDeploymentPreflight(facts, "https://identity.example.test", time.Unix(1, 0))
	second := BuildDeploymentPreflight(facts, "https://identity.example.test", time.Unix(2, 0))
	assert.Equal(t, first.ResourceRevision, second.ResourceRevision)
}

func TestDeploymentPreflightRejectsStatusThatDoesNotMatchChecks(t *testing.T) {
	preflight := BuildDeploymentPreflight(DeploymentPreflightFacts{}, "https://identity.example.test", time.Now())
	preflight.Status = PreflightReady
	require.ErrorContains(t, preflight.Validate(), "does not match checks")
}
