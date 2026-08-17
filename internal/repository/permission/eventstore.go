package permission

import "github.com/EonsofStupid/tessera/internal/eventstore"

func init() {
	eventstore.RegisterFilterEventMapper(AggregateType, AddedType, eventstore.GenericEventMapper[AddedEvent])
	eventstore.RegisterFilterEventMapper(AggregateType, RemovedType, eventstore.GenericEventMapper[RemovedEvent])
}
