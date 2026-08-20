package management

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/EonsofStupid/tessera/backend/v1/domain"
)

// CapabilityService publishes the safe development discovery document until a
// signed bundle manifest and its conformance evidence are installed. It is
// deliberately conservative: process health or inherited code is never
// promoted into an "operational" product claim.
type CapabilityService struct {
	clock Clock
}

func NewCapabilityService(clock Clock) *CapabilityService {
	if clock == nil {
		clock = time.Now
	}
	return &CapabilityService{clock: clock}
}

func (s *CapabilityService) Get(context.Context) (domain.CapabilityDiscovery, error) {
	observedAt := s.clock().UTC()
	components := []domain.ComponentCompatibility{
		{Role: domain.ComponentTessera, Version: "workspace", APIMajor: 1, State: domain.CompatibilityUnknown, Reason: "development_bundle_unattested", ObservedAt: observedAt},
		{Role: domain.ComponentTesseraOperator, State: domain.CompatibilityNotPresent, Reason: "operator_not_connected", ObservedAt: observedAt},
		{Role: domain.ComponentClickHouse, State: domain.CompatibilityNotPresent, Reason: "clickhouse_not_connected", ObservedAt: observedAt},
		{Role: domain.ComponentVaultix, State: domain.CompatibilityNotPresent, Reason: "vaultix_not_connected", ObservedAt: observedAt},
		{Role: domain.ComponentZuul, State: domain.CompatibilityNotPresent, Reason: "zuul_not_connected", ObservedAt: observedAt},
		{Role: domain.ComponentShippinAdapter, State: domain.CompatibilityNotPresent, Reason: "shippin_adapter_not_connected", ObservedAt: observedAt},
	}
	capabilities := developmentCapabilityFacts()
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

func developmentCapabilityFacts() []domain.CapabilityFact {
	preview := func(id, reason string, required ...domain.ComponentRole) domain.CapabilityFact {
		return domain.CapabilityFact{
			ID:                 id,
			Status:             domain.CapabilityPreview,
			Exposure:           domain.UIExposureDisabled,
			Reason:             reason,
			RequiredComponents: required,
			OperationKinds:     []domain.OperationKind{},
		}
	}
	return []domain.CapabilityFact{
		preview(domain.CapabilityIDOverview, "awaiting_conformance_proof", domain.ComponentTessera),
		preview(domain.CapabilityIDGuidedSetup, "operator_not_connected", domain.ComponentTessera, domain.ComponentTesseraOperator),
		preview(domain.CapabilityIDDeploymentOperations, "operator_not_connected", domain.ComponentTessera, domain.ComponentTesseraOperator),
		preview(domain.CapabilityIDAnalyticsOLAP, "clickhouse_not_connected", domain.ComponentTessera, domain.ComponentClickHouse),
		preview(domain.CapabilityIDUpstreamOIDC, "awaiting_conformance_proof", domain.ComponentTessera),
		preview(domain.CapabilityIDUpstreamSAML, "awaiting_conformance_proof", domain.ComponentTessera),
		preview(domain.CapabilityIDDownstreamOIDC, "awaiting_conformance_proof", domain.ComponentTessera),
		preview(domain.CapabilityIDDownstreamSAML, "awaiting_conformance_proof", domain.ComponentTessera),
		preview(domain.CapabilityIDLDAPOutbound, "ldap_outbound_conformance_pending", domain.ComponentTessera),
		preview(domain.CapabilityIDLDAPInbound, "ldap_inbound_conformance_pending", domain.ComponentTessera),
		preview(domain.CapabilityIDForwardAuth, "forward_auth_conformance_pending", domain.ComponentTessera),
		preview(domain.CapabilityIDIdentityAwareProxy, "proxy_conformance_pending", domain.ComponentTessera, domain.ComponentZuul),
		preview(domain.CapabilityIDVisualFlowEngine, "visual_flow_editor_pending", domain.ComponentTessera),
		preview(domain.CapabilityIDVaultixSecretCustody, "vaultix_not_connected", domain.ComponentTessera, domain.ComponentVaultix),
	}
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
