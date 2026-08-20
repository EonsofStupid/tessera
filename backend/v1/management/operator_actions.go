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

type OperatorActionService struct {
	capabilities *CapabilityService
	clock        Clock
}

func NewOperatorActionService(capabilities *CapabilityService, clock Clock) *OperatorActionService {
	if clock == nil {
		clock = time.Now
	}
	return &OperatorActionService{capabilities: capabilities, clock: clock}
}

func (s *OperatorActionService) Get(ctx context.Context) (domain.OperatorActionCatalog, error) {
	if s == nil || s.capabilities == nil {
		return domain.OperatorActionCatalog{}, fmt.Errorf("capability discovery is unavailable")
	}
	discovery, err := s.capabilities.Get(ctx)
	if err != nil {
		return domain.OperatorActionCatalog{}, err
	}
	actions := []domain.OperatorAction{
		plannedAction("action.application_plan_create", "Plan an application", "Creates an immutable application plan for operator review.", "/tessera/v1/applications:plan", "tessera.applications.plan", domain.CapabilityIDDownstreamOIDC),
		plannedAction("action.provider_plan_create", "Plan an identity provider", "Creates an immutable federation plan without changing live trust.", "/tessera/v1/providers:plan", "tessera.providers.plan", domain.CapabilityIDUpstreamOIDC),
		plannedAction("action.flow_plan_publish", "Plan a flow publication", "Validates a flow graph and prepares a reviewed publication plan.", "/tessera/v1/flows:plan", "tessera.flows.plan", domain.CapabilityIDVisualFlowEngine),
		plannedAction("action.deployment_plan_operation", "Plan a deployment operation", "Prepares a reviewed install, backup, restore, upgrade or rotation operation.", "/tessera/v1/deployment:plan", "tessera.deployment.plan", domain.CapabilityIDDeploymentOperations),
	}
	for index := range actions {
		resolution := domain.ResolveCapability(discovery, []uint32{1}, actions[index].CapabilityID)
		actions[index].Exposure = resolution.Exposure
		actions[index].Reason = resolution.Reason
		if actions[index].Exposure != domain.UIExposureEnabled && actions[index].Reason == "" {
			actions[index].Reason = "capability_not_operational"
		}
	}
	revisionSeed, err := json.Marshal(actions)
	if err != nil {
		return domain.OperatorActionCatalog{}, err
	}
	digest := sha256.Sum256(revisionSeed)
	catalog := domain.OperatorActionCatalog{
		SchemaVersion: 1, ResourceRevision: "sha256:" + hex.EncodeToString(digest[:]),
		ObservedAt: s.clock().UTC(), Actions: actions,
	}
	if err := catalog.Validate(); err != nil {
		return domain.OperatorActionCatalog{}, fmt.Errorf("assembled operator action catalog is invalid: %w", err)
	}
	return catalog, nil
}

func plannedAction(id, title, consequence, href, permission, capability string) domain.OperatorAction {
	return domain.OperatorAction{
		ID: id, Title: title, Consequence: consequence, Stage: domain.OperatorActionPlan,
		Method: "POST", Href: href, IntentSchema: json.RawMessage(`{"type":"object","additionalProperties":false}`),
		RequiredPermissions: []string{permission}, CapabilityID: capability,
		Exposure: domain.UIExposureDisabled, Reason: "capability_not_operational",
		SeedSuggestions: []domain.OperatorSuggestion{},
	}
}
