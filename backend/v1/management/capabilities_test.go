package management

import (
	"context"
	"testing"
	"time"

	"github.com/EonsofStupid/tessera/backend/v1/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilityServicePublishesEveryNamedGapWithoutFalseAvailability(t *testing.T) {
	now := time.Date(2026, time.August, 20, 2, 0, 0, 0, time.UTC)
	discovery, err := NewCapabilityService(func() time.Time { return now }).Get(context.Background())
	require.NoError(t, err)
	require.NoError(t, discovery.Validate())
	assert.Equal(t, now, discovery.ObservedAt)

	want := map[string]bool{
		domain.CapabilityIDLDAPOutbound:         false,
		domain.CapabilityIDLDAPInbound:          false,
		domain.CapabilityIDForwardAuth:          false,
		domain.CapabilityIDIdentityAwareProxy:   false,
		domain.CapabilityIDVisualFlowEngine:     false,
		domain.CapabilityIDVaultixSecretCustody: false,
	}
	for _, capability := range discovery.Capabilities {
		if _, tracked := want[capability.ID]; tracked {
			want[capability.ID] = true
			assert.Equal(t, domain.CapabilityPreview, capability.Status)
			assert.Equal(t, domain.UIExposureDisabled, capability.Exposure)
			assert.Nil(t, capability.Proof)
		}
	}
	for id, found := range want {
		assert.True(t, found, "missing capability %s", id)
	}
}

func TestCapabilityServiceRevisionDoesNotChangeWithObservationTime(t *testing.T) {
	first, err := NewCapabilityService(func() time.Time { return time.Unix(1, 0) }).Get(context.Background())
	require.NoError(t, err)
	second, err := NewCapabilityService(func() time.Time { return time.Unix(2, 0) }).Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, first.ResourceRevision, second.ResourceRevision)
}
