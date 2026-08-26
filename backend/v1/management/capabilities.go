package management

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shippinAI/nomen/backend/v1/domain"
)

// CapabilityService publishes the safe development discovery document until a
// signed bundle manifest and its conformance evidence are installed. It is
// deliberately conservative: process health or inherited code is never
// promoted into an "operational" product claim.
type CapabilityService struct {
	clock   Clock
	edition string
}

func NewCapabilityService(clock Clock, opts ...func(*CapabilityService)) *CapabilityService {
	if clock == nil {
		clock = time.Now
	}
	service := &CapabilityService{clock: clock, edition: domain.EditionPublic}
	for _, opt := range opts {
		opt(service)
	}
	service.edition = domain.NormalizeEdition(service.edition)
	return service
}

func WithEdition(edition string) func(*CapabilityService) {
	return func(s *CapabilityService) {
		s.edition = edition
	}
}

func (s *CapabilityService) Get(context.Context) (domain.CapabilityDiscovery, error) {
	observedAt := s.clock().UTC()
	clickhouseReason, vaultReason, meshReason := "clickhouse_not_connected", "vaultix_not_connected", "zuul_not_connected"
	if s.edition == domain.EditionPublic {
		clickhouseReason, vaultReason, meshReason = domain.ReasonEditionPublicWithheld, domain.ReasonEditionPublicWithheld, domain.ReasonEditionPublicWithheld
	}
	components := []domain.ComponentCompatibility{
		{Role: domain.ComponentNomen, Version: domain.ProductVersion, APIMajor: 1, State: domain.CompatibilityUnknown, Reason: "development_bundle_unattested", ObservedAt: observedAt},
		{Role: domain.ComponentNomenOperator, State: domain.CompatibilityNotPresent, Reason: "operator_not_connected", ObservedAt: observedAt},
		{Role: domain.ComponentClickHouse, State: domain.CompatibilityNotPresent, Reason: clickhouseReason, ObservedAt: observedAt},
		{Role: domain.ComponentVaultix, State: domain.CompatibilityNotPresent, Reason: vaultReason, ObservedAt: observedAt},
		{Role: domain.ComponentZuul, State: domain.CompatibilityNotPresent, Reason: meshReason, ObservedAt: observedAt},
		{Role: domain.ComponentShippinAdapter, State: domain.CompatibilityNotPresent, Reason: "shippin_adapter_not_connected", ObservedAt: observedAt},
	}
	capabilities, err := developmentCapabilityFacts(s.edition)
	if err != nil {
		return domain.CapabilityDiscovery{}, err
	}
	revision, err := capabilityRevision(components, capabilities)
	if err != nil {
		return domain.CapabilityDiscovery{}, err
	}
	discovery := domain.CapabilityDiscovery{
		SchemaVersion:    1,
		ResourceRevision: revision,
		ObservedAt:       observedAt,
		Components:       components,
		Capabilities:     capabilities,
	}
	if err := discovery.Validate(); err != nil {
		return domain.CapabilityDiscovery{}, fmt.Errorf("assembled capability discovery is invalid: %w", err)
	}
	return discovery, nil
}

func developmentCapabilityFacts(edition string) ([]domain.CapabilityFact, error) {
	ledger, err := domain.LoadCapabilityLedger()
	if err != nil {
		return nil, fmt.Errorf("load capability ledger: %w", err)
	}
	entries := ledger.Entries()
	facts := make([]domain.CapabilityFact, 0, len(entries))
	for _, entry := range entries {
		status := entry.CurrentStatus
		exposure := domain.UIExposureDisabled
		reason := "awaiting_release_bound_evidence"
		if edition == domain.EditionPublic && domain.PublicWithheldCapability(entry.ID) {
			status = domain.CapabilityUnsupported
			exposure = domain.UIExposureHidden
			reason = domain.ReasonEditionPublicWithheld
		}
		facts = append(facts, domain.CapabilityFact{
			ID:                 entry.ID,
			Status:             status,
			Exposure:           exposure,
			Reason:             reason,
			RequiredComponents: entry.RequiredComponents,
			OperationKinds:     []domain.OperationKind{},
		})
	}
	return facts, nil
}

func capabilityRevision(components []domain.ComponentCompatibility, capabilities []domain.CapabilityFact) (string, error) {
	type componentRevision struct {
		Role  domain.ComponentRole      `json:"role"`
		State domain.CompatibilityState `json:"state"`
	}
	seed := struct {
		Components   []componentRevision     `json:"components"`
		Capabilities []domain.CapabilityFact `json:"capabilities"`
	}{Capabilities: capabilities}
	for _, component := range components {
		seed.Components = append(seed.Components, componentRevision{Role: component.Role, State: component.State})
	}
	encoded, err := json.Marshal(seed)
	if err != nil {
		return "", fmt.Errorf("encode capability revision: %w", err)
	}
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}
