package setup

import (
	"context"

	"github.com/EonsofStupid/tessera/internal/api/authz"
	"github.com/EonsofStupid/tessera/internal/eventstore"
	"github.com/EonsofStupid/tessera/internal/query/projection"
	"github.com/EonsofStupid/tessera/internal/repository/instance"
)

type FillMembershipFields struct {
	eventstore *eventstore.Eventstore
}

func (mig *FillMembershipFields) Execute(ctx context.Context, _ eventstore.Event) error {
	instances, err := mig.eventstore.InstanceIDs(
		ctx,
		eventstore.NewSearchQueryBuilder(eventstore.ColumnsInstanceIDs).
			OrderDesc().
			AddQuery().
			AggregateTypes("instance").
			EventTypes(instance.InstanceAddedEventType).
			Builder().ExcludeAggregateIDs().
			AggregateTypes("instance").
			EventTypes(instance.InstanceRemovedEventType).
			Builder(),
	)
	if err != nil {
		return err
	}
	for _, instance := range instances {
		ctx := authz.WithInstanceID(ctx, instance)
		if err := projection.MembershipFields.Trigger(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (mig *FillMembershipFields) String() string {
	return "47_fill_membership_fields"
}
