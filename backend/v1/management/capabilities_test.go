package management

import (
	"context"
	"testing"
	"time"

	"github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilityServicePublishesEveryLedgerCapabilityWithoutFalseAvailability(t *testing.T) {
	now := time.Date(2026, time.August, 20, 2, 0, 0, 0, time.UTC)
	discovery, err := NewCapabilityService(func() time.Time { return now }).Get(context.Background())
	require.NoError(t, err)
	require.NoError(t, discovery.Validate())
	assert.Equal(t, now, discovery.ObservedAt)

	ledger, err := domain.LoadCapabilityLedger()
	require.NoError(t, err)
	want := make(map[string]bool)
	for _, entry := range ledger.Entries() {
		want[entry.ID] = false
	}
	for _, capability := range discovery.Capabilities {
		_, tracked := want[capability.ID]
		assert.True(t, tracked, "discovery contains capability absent from ledger: %s", capability.ID)
		want[capability.ID] = true
		if domain.PublicWithheldCapability(capability.ID) {
			assert.Equal(t, domain.CapabilityUnsupported, capability.Status)
			assert.Equal(t, domain.UIExposureHidden, capability.Exposure)
			assert.Equal(t, domain.ReasonEditionPublicWithheld, capability.Reason)
		} else {
			assert.Equal(t, domain.CapabilityPreview, capability.Status)
			assert.Equal(t, domain.UIExposureDisabled, capability.Exposure)
		}
		assert.Nil(t, capability.Proof)
	}
	for id, found := range want {
		assert.True(t, found, "missing capability %s", id)
	}
}

func TestCapabilityServiceHidesVaultOnPublicAndKeepsWireRoles(t *testing.T) {
	discovery, err := NewCapabilityService(time.Now, WithEdition(domain.EditionPublic)).Get(context.Background())
	require.NoError(t, err)
	roles := map[domain.ComponentRole]string{}
	for _, component := range discovery.Components {
		roles[component.Role] = component.Reason
	}
	assert.Equal(t, domain.ReasonEditionPublicWithheld, roles[domain.ComponentVaultix])
	assert.Equal(t, domain.ReasonEditionPublicWithheld, roles[domain.ComponentZuul])
	enterprise, err := NewCapabilityService(time.Now, WithEdition(domain.EditionEnterprise)).Get(context.Background())
	require.NoError(t, err)
	for _, capability := range enterprise.Capabilities {
		if capability.ID == domain.CapabilityIDVaultixSecretCustody {
			assert.Equal(t, domain.UIExposureDisabled, capability.Exposure)
			assert.Equal(t, "awaiting_release_bound_evidence", capability.Reason)
		}
	}
}

func TestCapabilityServiceRevisionDoesNotChangeWithObservationTime(t *testing.T) {
	first, err := NewCapabilityService(func() time.Time { return time.Unix(1, 0) }).Get(context.Background())
	require.NoError(t, err)
	second, err := NewCapabilityService(func() time.Time { return time.Unix(2, 0) }).Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, first.ResourceRevision, second.ResourceRevision)
}
