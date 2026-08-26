package target

import "github.com/shippinAI/nomen/internal/eventstore"

const (
	AggregateType    = "target"
	AggregateVersion = "v1"
)

func NewAggregate(aggregateID, instanceID string) *eventstore.Aggregate {
	return &eventstore.Aggregate{
		ID:            aggregateID,
		Type:          AggregateType,
		ResourceOwner: instanceID,
		InstanceID:    instanceID,
		Version:       AggregateVersion,
	}
}
