package restrictions

import (
	"github.com/EonsofStupid/tessera/internal/eventstore"
)

func init() {
	eventstore.RegisterFilterEventMapper(AggregateType, SetEventType, SetEventMapper)
}
