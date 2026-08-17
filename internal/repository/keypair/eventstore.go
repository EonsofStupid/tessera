package keypair

import (
	"github.com/EonsofStupid/tessera/internal/eventstore"
)

func init() {
	eventstore.RegisterFilterEventMapper(AggregateType, AddedEventType, AddedEventMapper)
	eventstore.RegisterFilterEventMapper(AggregateType, AddedCertificateEventType, AddedCertificateEventMapper)
}
