package keypair

import (
	"github.com/shippinAI/nomen/internal/eventstore"
)

func init() {
	eventstore.RegisterFilterEventMapper(AggregateType, AddedEventType, AddedEventMapper)
	eventstore.RegisterFilterEventMapper(AggregateType, AddedCertificateEventType, AddedCertificateEventMapper)
}
