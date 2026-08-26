package milestone

import (
	"context"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/eventstore"
)

const (
	AggregateType    = "milestone"
	AggregateVersion = "v2"
)

type Aggregate struct {
	eventstore.Aggregate
}

func NewAggregate(ctx context.Context) *Aggregate {
	return NewInstanceAggregate(authz.GetInstance(ctx).InstanceID())
}

func NewInstanceAggregate(instanceID string) *Aggregate {
	return &Aggregate{
		Aggregate: eventstore.Aggregate{
			Type:          AggregateType,
			Version:       AggregateVersion,
			ID:            instanceID,
			ResourceOwner: instanceID,
			InstanceID:    instanceID,
		},
	}
}
