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

type overviewSourceStub struct {
	facts domain.OverviewFacts
	err   error
}

func (s overviewSourceStub) Snapshot(context.Context, string) (domain.OverviewFacts, error) {
	return s.facts, s.err
}

type keyCounterStub struct {
	count uint32
	err   error
}

func (s keyCounterStub) ActiveSigningKeys(context.Context) (uint32, error) {
	return s.count, s.err
}

func TestOverviewServiceBuildsValidatedProjection(t *testing.T) {
	observed := time.Date(2026, 8, 19, 1, 2, 3, 0, time.UTC)
	service := NewOverviewService(
		overviewSourceStub{facts: domain.OverviewFacts{HumanSeats: 2, AgentSeats: 1, WorkspaceAttachments: 3, Flows: 1}},
		keyCounterStub{count: 1},
		func() time.Time { return observed },
	)

	overview, err := service.Get(context.Background(), "instance-1", "https://id.shippin.ai")
	require.NoError(t, err)
	assert.Equal(t, observed, overview.ObservedAt)
	assert.Equal(t, domain.OverviewReady, overview.Readiness.Status)
}

func TestOverviewServiceDoesNotTurnSourceFailureIntoZeros(t *testing.T) {
	service := NewOverviewService(
		overviewSourceStub{err: errors.New("database offline")},
		keyCounterStub{count: 1},
		time.Now,
	)

	_, err := service.Get(context.Background(), "instance-1", "https://id.shippin.ai")
	require.ErrorContains(t, err, "overview facts unavailable")
}
