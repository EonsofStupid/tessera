package permission

import "github.com/shippinAI/nomen/internal/eventstore"

func init() {
	eventstore.RegisterFilterEventMapper(AggregateType, AddedType, eventstore.GenericEventMapper[AddedEvent])
	eventstore.RegisterFilterEventMapper(AggregateType, RemovedType, eventstore.GenericEventMapper[RemovedEvent])
}
