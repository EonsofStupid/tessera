package domain

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

type ComponentRole string

const (
	ComponentTessera         ComponentRole = "tessera"
	ComponentTesseraOperator ComponentRole = "tessera_operator"
	ComponentClickHouse      ComponentRole = "clickhouse"
	ComponentVaultix         ComponentRole = "vaultix"
	ComponentZuul            ComponentRole = "zuul"
	ComponentShippinAdapter  ComponentRole = "shippin_adapter"
)

func (r ComponentRole) Valid() bool {
	return r == ComponentTessera ||
		r == ComponentTesseraOperator ||
		r == ComponentClickHouse ||
		r == ComponentVaultix ||
		r == ComponentZuul ||
		r == ComponentShippinAdapter
}

const (
	CapabilityIDOverview             = "overview"
	CapabilityIDGuidedSetup          = "guided_setup"
	CapabilityIDDeploymentOperations = "deployment_operations"
	CapabilityIDAnalyticsOLAP        = "analytics_olap"
	CapabilityIDUpstreamOIDC         = "upstream_oidc"
	CapabilityIDUpstreamSAML         = "upstream_saml"
	CapabilityIDDownstreamOIDC       = "downstream_oidc"
	CapabilityIDDownstreamSAML       = "downstream_saml"
	CapabilityIDLDAPOutbound         = "ldap_outbound"
	CapabilityIDLDAPInbound          = "ldap_inbound"
	CapabilityIDForwardAuth          = "forward_auth"
	CapabilityIDIdentityAwareProxy   = "identity_aware_proxy"
	CapabilityIDVisualFlowEngine     = "visual_flow_engine"
	CapabilityIDVaultixSecretCustody = "vaultix_secret_custody"
)

type CompatibilityState string

const (
	CompatibilityCompatible   CompatibilityState = "compatible"
	CompatibilityIncompatible CompatibilityState = "incompatible"
	CompatibilityNotPresent   CompatibilityState = "not_present"
	CompatibilityUnknown      CompatibilityState = "unknown"
)

func (s CompatibilityState) Valid() bool {
	return s == CompatibilityCompatible || s == CompatibilityIncompatible || s == CompatibilityNotPresent || s == CompatibilityUnknown
}

type CapabilityStatus string

const (
	CapabilityUnsupported CapabilityStatus = "unsupported"
	CapabilityPreview     CapabilityStatus = "preview"
	CapabilityOperational CapabilityStatus = "operational"
	CapabilityDegraded    CapabilityStatus = "degraded"
)

func (s CapabilityStatus) Valid() bool {
	return s == CapabilityUnsupported || s == CapabilityPreview || s == CapabilityOperational || s == CapabilityDegraded
}

type UIExposure string

const (
	UIExposureHidden   UIExposure = "hidden"
	UIExposureDisabled UIExposure = "disabled"
	UIExposureEnabled  UIExposure = "enabled"
)

func (e UIExposure) Valid() bool {
	return e == UIExposureHidden || e == UIExposureDisabled || e == UIExposureEnabled
}

type ConformanceResult string

const (
	ConformancePassed ConformanceResult = "passed"
	ConformanceFailed ConformanceResult = "failed"
)

func (r ConformanceResult) Valid() bool {
	return r == ConformancePassed || r == ConformanceFailed
}

type ComponentCompatibility struct {
	Role       ComponentRole      `json:"role"`
	Version    string             `json:"version"`
	APIMajor   uint32             `json:"api_major"`
	State      CompatibilityState `json:"state"`
	Reason     string             `json:"reason,omitempty"`
	ObservedAt time.Time          `json:"observed_at"`
}

type CapabilityFact struct {
	ID                 string           `json:"id"`
	Status             CapabilityStatus `json:"status"`
	Exposure           UIExposure       `json:"exposure"`
	Reason             string           `json:"reason,omitempty"`
	RequiredComponents []ComponentRole  `json:"required_components"`
	OperationKinds     []OperationKind  `json:"operation_kinds"`
	Proof              *CapabilityProof `json:"proof,omitempty"`
}

type CapabilityProof struct {
	ConformanceID        string            `json:"conformance_id"`
	BundleManifestDigest string            `json:"bundle_manifest_digest"`
	Result               ConformanceResult `json:"result"`
	VerifiedAt           time.Time         `json:"verified_at"`
	EvidenceDigest       string            `json:"evidence_digest"`
}

type CapabilityDiscovery struct {
	SchemaVersion        uint32                   `json:"schema_version"`
	ResourceRevision     string                   `json:"resource_revision"`
	BundleManifestDigest string                   `json:"bundle_manifest_digest,omitempty"`
	ObservedAt           time.Time                `json:"observed_at"`
	Components           []ComponentCompatibility `json:"components"`
	Capabilities         []CapabilityFact         `json:"capabilities"`
}

type CapabilityResolution struct {
	CapabilityID string     `json:"capability_id"`
	Exposure     UIExposure `json:"exposure"`
	Reason       string     `json:"reason,omitempty"`
}

func (d CapabilityDiscovery) Validate() error {
	if d.SchemaVersion != 1 {
		return fmt.Errorf("unsupported discovery schema version %d", d.SchemaVersion)
	}
	if strings.TrimSpace(d.ResourceRevision) == "" {
		return fmt.Errorf("resource_revision is required")
	}
	if d.ObservedAt.IsZero() {
		return fmt.Errorf("observed_at is required")
	}
	if d.BundleManifestDigest != "" && !validPlanDigest(d.BundleManifestDigest) {
		return fmt.Errorf("bundle_manifest_digest must be a lowercase sha256 digest")
	}
	if err := validateComponents(d.Components); err != nil {
		return err
	}
	return validateCapabilities(d.Capabilities, d.BundleManifestDigest)
}

func ResolveCapability(discovery CapabilityDiscovery, supportedSchemas []uint32, capabilityID string) CapabilityResolution {
	if !slices.Contains(supportedSchemas, discovery.SchemaVersion) {
		return CapabilityResolution{capabilityID, UIExposureDisabled, "schema_incompatible"}
	}
	if err := discovery.Validate(); err != nil {
		return CapabilityResolution{capabilityID, UIExposureDisabled, "invalid_discovery"}
	}

	var capability *CapabilityFact
	for index := range discovery.Capabilities {
		if discovery.Capabilities[index].ID == capabilityID {
			capability = &discovery.Capabilities[index]
			break
		}
	}
	if capability == nil {
		return CapabilityResolution{capabilityID, UIExposureHidden, "not_advertised"}
	}
	if capability.Exposure == UIExposureHidden {
		return CapabilityResolution{capabilityID, UIExposureHidden, capability.Reason}
	}

	for _, required := range capability.RequiredComponents {
		component, found := componentFor(discovery.Components, required)
		if !found || component.State != CompatibilityCompatible {
			return CapabilityResolution{capabilityID, UIExposureDisabled, "component_unavailable"}
		}
	}
	return CapabilityResolution{capabilityID, capability.Exposure, capability.Reason}
}

func validateComponents(components []ComponentCompatibility) error {
	seen := make(map[ComponentRole]struct{}, len(components))
	for _, component := range components {
		if !component.Role.Valid() {
			return fmt.Errorf("unknown component role %q", component.Role)
		}
		if _, duplicate := seen[component.Role]; duplicate {
			return fmt.Errorf("duplicate component role %q", component.Role)
		}
		seen[component.Role] = struct{}{}
		if component.ObservedAt.IsZero() || !component.State.Valid() {
			return fmt.Errorf("component %s has incomplete compatibility facts", component.Role)
		}
		if component.State != CompatibilityNotPresent && (strings.TrimSpace(component.Version) == "" || component.APIMajor == 0) {
			return fmt.Errorf("component %s has incomplete version facts", component.Role)
		}
		if component.State != CompatibilityCompatible && strings.TrimSpace(component.Reason) == "" {
			return fmt.Errorf("component %s requires a compatibility reason", component.Role)
		}
	}
	for _, mandatory := range []ComponentRole{ComponentTessera} {
		if _, ok := seen[mandatory]; !ok {
			return fmt.Errorf("mandatory component %s is missing", mandatory)
		}
	}
	return nil
}

func validateCapabilities(capabilities []CapabilityFact, bundleManifestDigest string) error {
	seen := make(map[string]struct{}, len(capabilities))
	for _, capability := range capabilities {
		if strings.TrimSpace(capability.ID) == "" || !capability.Status.Valid() || !capability.Exposure.Valid() {
			return fmt.Errorf("capability has incomplete identity or state")
		}
		if _, duplicate := seen[capability.ID]; duplicate {
			return fmt.Errorf("duplicate capability %q", capability.ID)
		}
		seen[capability.ID] = struct{}{}
		if capability.Exposure != UIExposureEnabled && strings.TrimSpace(capability.Reason) == "" {
			return fmt.Errorf("capability %s requires a reason when not enabled", capability.ID)
		}
		if capability.Exposure == UIExposureEnabled && (capability.Status == CapabilityUnsupported || capability.Status == CapabilityPreview) {
			return fmt.Errorf("capability %s cannot enable status %s", capability.ID, capability.Status)
		}
		if capability.Proof != nil {
			if err := validateCapabilityProof(capability.ID, *capability.Proof, bundleManifestDigest); err != nil {
				return err
			}
		}
		if capability.Status == CapabilityOperational || capability.Status == CapabilityDegraded {
			if capability.Proof == nil || capability.Proof.Result != ConformancePassed {
				return fmt.Errorf("capability %s requires passing conformance proof", capability.ID)
			}
		}
		componentSeen := make(map[ComponentRole]struct{}, len(capability.RequiredComponents))
		for _, role := range capability.RequiredComponents {
			if !role.Valid() {
				return fmt.Errorf("capability %s has unknown required component %q", capability.ID, role)
			}
			if _, duplicate := componentSeen[role]; duplicate {
				return fmt.Errorf("capability %s repeats required component %s", capability.ID, role)
			}
			componentSeen[role] = struct{}{}
		}
		operationSeen := make(map[OperationKind]struct{}, len(capability.OperationKinds))
		for _, kind := range capability.OperationKinds {
			if !kind.Valid() {
				return fmt.Errorf("capability %s has unknown operation kind %q", capability.ID, kind)
			}
			if _, duplicate := operationSeen[kind]; duplicate {
				return fmt.Errorf("capability %s repeats operation kind %s", capability.ID, kind)
			}
			operationSeen[kind] = struct{}{}
		}
	}
	return nil
}

func validateCapabilityProof(capabilityID string, proof CapabilityProof, bundleManifestDigest string) error {
	if strings.TrimSpace(proof.ConformanceID) == "" || proof.VerifiedAt.IsZero() || !proof.Result.Valid() {
		return fmt.Errorf("capability %s has incomplete conformance proof", capabilityID)
	}
	if !validPlanDigest(proof.BundleManifestDigest) || !validPlanDigest(proof.EvidenceDigest) {
		return fmt.Errorf("capability %s proof digests must be lowercase sha256 digests", capabilityID)
	}
	if bundleManifestDigest == "" || proof.BundleManifestDigest != bundleManifestDigest {
		return fmt.Errorf("capability %s proof does not match installed bundle", capabilityID)
	}
	return nil
}

func componentFor(components []ComponentCompatibility, role ComponentRole) (ComponentCompatibility, bool) {
	for _, component := range components {
		if component.Role == role {
			return component, true
		}
	}
	return ComponentCompatibility{}, false
}
