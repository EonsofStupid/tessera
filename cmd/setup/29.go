package setup

import (
	"context"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/eventstore"
	"github.com/shippinAI/nomen/internal/query/projection"
	"github.com/shippinAI/nomen/internal/repository/instance"
)

type FillFieldsForProjectGrant struct {
	eventstore *eventstore.Eventstore
}

func (mig *FillFieldsForProjectGrant) Execute(ctx context.Context, _ eventstore.Event) error {
	instances, err := mig.eventstore.InstanceIDs(
		ctx,
		eventstore.NewSearchQueryBuilder(eventstore.ColumnsInstanceIDs).
			OrderDesc().
			AddQuery().
			AggregateTypes("instance").
			EventTypes(instance.InstanceAddedEventType).
			Builder(),
	)
	if err != nil {
		return err
	}
	for _, instance := range instances {
		ctx := authz.WithInstanceID(ctx, instance)
		if err := projection.ProjectGrantFields.Trigger(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (mig *FillFieldsForProjectGrant) String() string {
	return "29_init_fields_for_project_grant"
}
