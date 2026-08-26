package samlsession

import (
	"github.com/shippinAI/nomen/internal/eventstore"
)

func init() {
	eventstore.RegisterFilterEventMapper(AggregateType, AddedType, eventstore.GenericEventMapper[AddedEvent])
	eventstore.RegisterFilterEventMapper(AggregateType, SAMLResponseAddedType, eventstore.GenericEventMapper[SAMLResponseAddedEvent])
	eventstore.RegisterFilterEventMapper(AggregateType, SAMLResponseRevokedType, eventstore.GenericEventMapper[SAMLResponseRevokedEvent])

}
