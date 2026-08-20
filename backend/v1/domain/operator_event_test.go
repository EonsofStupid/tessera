package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func validOperatorEvent(now time.Time) OperatorEvent {
	return OperatorEvent{
		SchemaVersion: OperatorEventSchemaVersion,
		EventID:       "9a5af3d5-5456-4266-8b24-c0b954db6819",
		SessionID:     "ce46aa0d-bc16-42e7-9975-1476ec5734f8",
		Sequence:      1,
		OccurredAt:    now,
		RouteID:       "route.federation",
		ControlID:     "control.provider_create",
		EventType:     OperatorEventControlActivated,
		Outcome:       OperatorOutcomeObserved,
	}
}

func TestOperatorEventBatchAcceptsSafeSemanticEvent(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	require.NoError(t, (OperatorEventBatch{Events: []OperatorEvent{validOperatorEvent(now)}}).Validate(now))
}

func TestOperatorEventBatchRejectsUnknownAttribute(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	event := validOperatorEvent(now)
	event.Attributes = map[string]string{"password": "must-not-be-recorded"}
	require.ErrorContains(t, (OperatorEventBatch{Events: []OperatorEvent{event}}).Validate(now), "not allowed")
}

func TestOperatorEventBatchRejectsDuplicateSequence(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	first := validOperatorEvent(now)
	second := first
	second.EventID = "44b3757f-7e78-4d11-b0f0-39a0f2f07354"
	require.ErrorContains(t, (OperatorEventBatch{Events: []OperatorEvent{first, second}}).Validate(now), "duplicate session sequence")
}
