package org

import (
	"github.com/shippinAI/nomen/internal/v2/eventstore"
	"github.com/shippinAI/nomen/internal/zerrors"
)

const DeactivatedType = eventTypePrefix + "deactivated"

type DeactivatedEvent eventstore.Event[eventstore.EmptyPayload]

var _ eventstore.TypeChecker = (*DeactivatedEvent)(nil)

// ActionType implements eventstore.Typer.
func (c *DeactivatedEvent) ActionType() string {
	return DeactivatedType
}

func DeactivatedEventFromStorage(event *eventstore.StorageEvent) (e *DeactivatedEvent, _ error) {
	if event.Type != e.ActionType() {
		return nil, zerrors.ThrowInvalidArgument(nil, "ORG-4zeWH", "Errors.Invalid.Event.Type")
	}

	return &DeactivatedEvent{
		StorageEvent: event,
	}, nil
}
