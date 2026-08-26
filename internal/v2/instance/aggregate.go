package instance

import "github.com/shippinAI/nomen/internal/repository/instance"

const (
	AggregateType   = string(instance.AggregateType)
	eventTypePrefix = AggregateType + "."
)
