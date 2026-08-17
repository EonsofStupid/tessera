package feature

import (
	"github.com/EonsofStupid/tessera/internal/eventstore"
)

const (
	eventTypePrefix = eventstore.EventType("feature.")
	setSuffix       = ".set"
)

const (
	AggregateType    = "feature"
	AggregateVersion = "v1"
)
