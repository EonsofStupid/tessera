package action

import "github.com/shippinAI/nomen/internal/eventstore"

func init() {
	eventstore.RegisterFilterEventMapper(AggregateType, AddedEventType, AddedEventMapper)
	eventstore.RegisterFilterEventMapper(AggregateType, ChangedEventType, ChangedEventMapper)
	eventstore.RegisterFilterEventMapper(AggregateType, DeactivatedEventType, DeactivatedEventMapper)
	eventstore.RegisterFilterEventMapper(AggregateType, ReactivatedEventType, ReactivatedEventMapper)
	eventstore.RegisterFilterEventMapper(AggregateType, RemovedEventType, RemovedEventMapper)
}
