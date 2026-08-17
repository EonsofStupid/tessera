package instance

import (
	"github.com/EonsofStupid/tessera/internal/v2/eventstore"
	"github.com/EonsofStupid/tessera/internal/zerrors"
)

const RemovedType = eventTypePrefix + "removed"

type RemovedEvent eventstore.Event[eventstore.EmptyPayload]

var _ eventstore.TypeChecker = (*RemovedEvent)(nil)

// ActionType implements eventstore.Typer.
func (c *RemovedEvent) ActionType() string {
	return RemovedType
}

func RemovedEventFromStorage(event *eventstore.StorageEvent) (e *RemovedEvent, _ error) {
	if event.Type != e.ActionType() {
		return nil, zerrors.ThrowInvalidArgument(nil, "INSTA-xppIg", "Errors.Invalid.Event.Type")
	}

	return &RemovedEvent{
		StorageEvent: event,
	}, nil
}
