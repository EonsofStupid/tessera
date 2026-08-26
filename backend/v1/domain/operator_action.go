package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type OperatorActionStage string

const (
	OperatorActionRead   OperatorActionStage = "read"
	OperatorActionPlan   OperatorActionStage = "plan"
	OperatorActionApply  OperatorActionStage = "apply"
	OperatorActionVerify OperatorActionStage = "verify"
	OperatorActionCancel OperatorActionStage = "cancel"
)

func (s OperatorActionStage) Valid() bool {
	return s == OperatorActionRead || s == OperatorActionPlan || s == OperatorActionApply || s == OperatorActionVerify || s == OperatorActionCancel
}

type OperatorSuggestion struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Value       string `json:"value,omitempty"`
	Source      string `json:"source"`
	Explanation string `json:"explanation"`
}

type OperatorAction struct {
	ID                  string               `json:"id"`
	Title               string               `json:"title"`
	Consequence         string               `json:"consequence"`
	Stage               OperatorActionStage  `json:"stage"`
	Method              string               `json:"method"`
	Href                string               `json:"href"`
	IntentSchema        json.RawMessage      `json:"intent_schema"`
	RequiredPermissions []string             `json:"required_permissions"`
	RequiredAssurance   string               `json:"required_assurance,omitempty"`
	CapabilityID        string               `json:"capability_id"`
	Exposure            UIExposure           `json:"exposure"`
	Reason              string               `json:"reason,omitempty"`
	Reversible          bool                 `json:"reversible"`
	SeedSuggestions     []OperatorSuggestion `json:"seed_suggestions"`
}

type OperatorActionCatalog struct {
	SchemaVersion    uint32           `json:"schema_version"`
	ResourceRevision string           `json:"resource_revision"`
	ObservedAt       time.Time        `json:"observed_at"`
	Actions          []OperatorAction `json:"actions"`
}

func (c OperatorActionCatalog) Validate() error {
	if c.SchemaVersion != 1 || strings.TrimSpace(c.ResourceRevision) == "" || c.ObservedAt.IsZero() {
		return fmt.Errorf("operator action catalog metadata is incomplete")
	}
	seen := make(map[string]struct{}, len(c.Actions))
	for index := range c.Actions {
		action := c.Actions[index]
		if !stableOperatorID.MatchString(action.ID) || strings.TrimSpace(action.Title) == "" || strings.TrimSpace(action.Consequence) == "" {
			return fmt.Errorf("action %d has incomplete identity", index)
		}
		if _, duplicate := seen[action.ID]; duplicate {
			return fmt.Errorf("duplicate action %q", action.ID)
		}
		seen[action.ID] = struct{}{}
		if !action.Stage.Valid() || !action.Exposure.Valid() || !strings.HasPrefix(action.Href, "/nomen/v1/") || !json.Valid(action.IntentSchema) {
			return fmt.Errorf("action %s has an invalid execution contract", action.ID)
		}
		if action.Exposure != UIExposureEnabled && strings.TrimSpace(action.Reason) == "" {
			return fmt.Errorf("action %s requires a disabled reason", action.ID)
		}
		if len(action.RequiredPermissions) == 0 || strings.TrimSpace(action.CapabilityID) == "" {
			return fmt.Errorf("action %s has incomplete authorization", action.ID)
		}
		for _, suggestion := range action.SeedSuggestions {
			if !stableOperatorID.MatchString(suggestion.ID) || suggestion.Label == "" || suggestion.Source == "" || suggestion.Explanation == "" {
				return fmt.Errorf("action %s has an invalid seed suggestion", action.ID)
			}
		}
	}
	return nil
}
