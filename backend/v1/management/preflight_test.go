package management

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type healthCheckerStub struct{ err error }

func (s healthCheckerStub) Health(context.Context) error { return s.err }

type signingKeyCounterStub struct {
	keys uint32
	err  error
}

func (s signingKeyCounterStub) ActiveSigningKeys(context.Context) (uint32, error) {
	return s.keys, s.err
}

type notificationCheckerStub struct {
	configured bool
	err        error
}

func (s notificationCheckerStub) Configured(context.Context, string) (bool, error) {
	return s.configured, s.err
}

func TestDeploymentPreflightServiceReturnsObservedFacts(t *testing.T) {
	now := time.Date(2026, time.August, 22, 2, 0, 0, 0, time.UTC)
	service := NewDeploymentPreflightService(healthCheckerStub{}, signingKeyCounterStub{keys: 1}, notificationCheckerStub{configured: true}, func() time.Time { return now })

	preflight, err := service.Get(context.Background(), "instance-1", "https://identity.example.test")
	require.NoError(t, err)
	assert.Equal(t, domain.PreflightReady, preflight.Status)
	assert.Equal(t, now, preflight.ObservedAt)
}

func TestDeploymentPreflightServiceSanitizesProbeFailureIntoChecks(t *testing.T) {
	probeError := errors.New("password=never-return-this")
	service := NewDeploymentPreflightService(healthCheckerStub{err: probeError}, signingKeyCounterStub{err: probeError}, notificationCheckerStub{err: probeError}, time.Now)

	preflight, err := service.Get(context.Background(), "instance-1", "https://identity.example.test")
	require.NoError(t, err)
	assert.Equal(t, domain.PreflightBlocked, preflight.Status)
	for _, check := range preflight.Checks {
		assert.NotContains(t, check.Summary, "never-return-this")
		assert.NotContains(t, check.Remediation, "never-return-this")
		assert.NotContains(t, check.DiagnosticRef, "never-return-this")
	}
}
