package instance

import "github.com/EonsofStupid/tessera/internal/repository/instance"

const (
	AggregateType   = string(instance.AggregateType)
	eventTypePrefix = AggregateType + "."
)
