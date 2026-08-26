package usergrant

import (
	"github.com/shippinAI/nomen/internal/eventstore"
)

func init() {
	eventstore.RegisterFilterEventMapper(AggregateType, UserGrantAddedType, UserGrantAddedEventMapper)
	eventstore.RegisterFilterEventMapper(AggregateType, UserGrantChangedType, UserGrantChangedEventMapper)
	eventstore.RegisterFilterEventMapper(AggregateType, UserGrantCascadeChangedType, UserGrantCascadeChangedEventMapper)
	eventstore.RegisterFilterEventMapper(AggregateType, UserGrantRemovedType, UserGrantRemovedEventMapper)
	eventstore.RegisterFilterEventMapper(AggregateType, UserGrantCascadeRemovedType, UserGrantCascadeRemovedEventMapper)
	eventstore.RegisterFilterEventMapper(AggregateType, UserGrantDeactivatedType, UserGrantDeactivatedEventMapper)
	eventstore.RegisterFilterEventMapper(AggregateType, UserGrantReactivatedType, UserGrantReactivatedEventMapper)
}
