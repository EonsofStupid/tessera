package feature

import (
	"github.com/shippinAI/nomen/internal/eventstore"
)

func init() {
	eventstore.RegisterFilterEventMapper(AggregateType, DefaultLoginInstanceEventType, eventstore.GenericEventMapper[SetEvent[Boolean]])
}
