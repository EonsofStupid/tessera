package feature

import (
	"github.com/EonsofStupid/tessera/internal/eventstore"
)

func init() {
	eventstore.RegisterFilterEventMapper(AggregateType, DefaultLoginInstanceEventType, eventstore.GenericEventMapper[SetEvent[Boolean]])
}
