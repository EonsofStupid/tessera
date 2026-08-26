package limits

import (
	"github.com/shippinAI/nomen/internal/eventstore"
)

func init() {
	eventstore.RegisterFilterEventMapper(AggregateType, SetEventType, SetEventMapper)
	eventstore.RegisterFilterEventMapper(AggregateType, ResetEventType, ResetEventMapper)
}
