package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	OperatorEventSchemaVersion uint32 = 1
	OperatorEventBatchLimit           = 50
)

type OperatorEventType string

const (
	OperatorEventRouteOpened        OperatorEventType = "route_opened"
	OperatorEventControlActivated   OperatorEventType = "control_activated"
	OperatorEventSuggestionAccepted OperatorEventType = "suggestion_accepted"
	OperatorEventGuideAdvanced      OperatorEventType = "guide_advanced"
	OperatorEventActionResult       OperatorEventType = "action_result"
)

func (t OperatorEventType) Valid() bool {
	return t == OperatorEventRouteOpened ||
		t == OperatorEventControlActivated ||
		t == OperatorEventSuggestionAccepted ||
		t == OperatorEventGuideAdvanced ||
		t == OperatorEventActionResult
}

type OperatorEventOutcome string

const (
	OperatorOutcomeObserved OperatorEventOutcome = "observed"
	OperatorOutcomeAccepted OperatorEventOutcome = "accepted"
	OperatorOutcomeRefused  OperatorEventOutcome = "refused"
	OperatorOutcomeFailed   OperatorEventOutcome = "failed"
)

func (o OperatorEventOutcome) Valid() bool {
	return o == "" || o == OperatorOutcomeObserved || o == OperatorOutcomeAccepted || o == OperatorOutcomeRefused || o == OperatorOutcomeFailed
}

var stableOperatorID = regexp.MustCompile(`^[a-z][a-z0-9]*(?:[._][a-z0-9_]+)+$`)

var operatorAttributeAllowlist = map[string]struct{}{
	"capability_status":  {},
	"resource_revision":  {},
	"selected_tab":       {},
	"deployment_profile": {},
}

type OperatorEvent struct {
	SchemaVersion    uint32               `json:"schema_version"`
	EventID          string               `json:"event_id"`
	SessionID        string               `json:"session_id"`
	Sequence         uint64               `json:"sequence"`
	OccurredAt       time.Time            `json:"occurred_at"`
	RouteID          string               `json:"route_id"`
	ControlID        string               `json:"control_id,omitempty"`
	EventType        OperatorEventType    `json:"event_type"`
	ActionID         string               `json:"action_id,omitempty"`
	ResourceRevision string               `json:"resource_revision,omitempty"`
	CorrelationID    string               `json:"correlation_id,omitempty"`
	Outcome          OperatorEventOutcome `json:"outcome,omitempty"`
	Attributes       map[string]string    `json:"attributes,omitempty"`
}

type OperatorEventBatch struct {
	Events []OperatorEvent `json:"events"`
}

type OperatorActor struct {
	InstanceID string
	TenantID   string
	ActorID    string
	AgentID    string
}

func (b OperatorEventBatch) Validate(now time.Time) error {
	if len(b.Events) == 0 || len(b.Events) > OperatorEventBatchLimit {
		return fmt.Errorf("operator event batch must contain 1-%d events", OperatorEventBatchLimit)
	}
	seenIDs := make(map[string]struct{}, len(b.Events))
	seenSequence := make(map[string]struct{}, len(b.Events))
	for index := range b.Events {
		if err := b.Events[index].Validate(now); err != nil {
			return fmt.Errorf("events[%d]: %w", index, err)
		}
		if _, duplicate := seenIDs[b.Events[index].EventID]; duplicate {
			return fmt.Errorf("events[%d]: duplicate event_id", index)
		}
		seenIDs[b.Events[index].EventID] = struct{}{}
		sequenceKey := fmt.Sprintf("%s:%d", b.Events[index].SessionID, b.Events[index].Sequence)
		if _, duplicate := seenSequence[sequenceKey]; duplicate {
			return fmt.Errorf("events[%d]: duplicate session sequence", index)
		}
		seenSequence[sequenceKey] = struct{}{}
	}
	return nil
}

func (e OperatorEvent) Validate(now time.Time) error {
	if e.SchemaVersion != OperatorEventSchemaVersion {
		return fmt.Errorf("unsupported schema_version")
	}
	if _, err := uuid.Parse(e.EventID); err != nil {
		return fmt.Errorf("event_id must be a UUID")
	}
	if _, err := uuid.Parse(e.SessionID); err != nil {
		return fmt.Errorf("session_id must be a UUID")
	}
	if e.Sequence == 0 {
		return fmt.Errorf("sequence must be positive")
	}
	if e.OccurredAt.IsZero() || e.OccurredAt.Before(now.Add(-24*time.Hour)) || e.OccurredAt.After(now.Add(5*time.Minute)) {
		return fmt.Errorf("occurred_at is outside the accepted window")
	}
	if !stableOperatorID.MatchString(e.RouteID) {
		return fmt.Errorf("route_id is not a stable identifier")
	}
	if !e.EventType.Valid() || !e.Outcome.Valid() {
		return fmt.Errorf("event type or outcome is invalid")
	}
	if e.EventType != OperatorEventRouteOpened && !stableOperatorID.MatchString(e.ControlID) {
		return fmt.Errorf("control_id is required")
	}
	if e.ActionID != "" && (!stableOperatorID.MatchString(e.ActionID) || len(e.ActionID) > 160) {
		return fmt.Errorf("action identifier is invalid")
	}
	if e.CorrelationID != "" {
		_, uuidError := uuid.Parse(e.CorrelationID)
		if uuidError != nil && !stableOperatorID.MatchString(e.CorrelationID) {
			return fmt.Errorf("correlation identifier is invalid")
		}
	}
	if len(e.ResourceRevision) > 96 {
		return fmt.Errorf("resource_revision is too long")
	}
	for key, value := range e.Attributes {
		if _, allowed := operatorAttributeAllowlist[key]; !allowed {
			return fmt.Errorf("attribute %q is not allowed", key)
		}
		if strings.TrimSpace(value) == "" || len(value) > 160 {
			return fmt.Errorf("attribute %q has an invalid value", key)
		}
	}
	return nil
}

func (a OperatorActor) Validate() error {
	if strings.TrimSpace(a.InstanceID) == "" || strings.TrimSpace(a.TenantID) == "" || strings.TrimSpace(a.ActorID) == "" {
		return fmt.Errorf("operator actor requires instance, tenant and actor")
	}
	return nil
}
