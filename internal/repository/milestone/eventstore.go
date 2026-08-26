package milestone

import (
	"github.com/shippinAI/nomen/internal/eventstore"
)

var (
	ReachedEventMapper = eventstore.GenericEventMapper[ReachedEvent]
	PushedEventMapper  = eventstore.GenericEventMapper[PushedEvent]
)

func init() {
	eventstore.RegisterFilterEventMapper(AggregateType, ReachedEventType, ReachedEventMapper)
	eventstore.RegisterFilterEventMapper(AggregateType, PushedEventType, PushedEventMapper)
}
