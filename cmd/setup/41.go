package setup

import (
	"context"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/eventstore"
	"github.com/shippinAI/nomen/internal/query/projection"
	"github.com/shippinAI/nomen/internal/repository/instance"
)

type FillFieldsForInstanceDomains struct {
	eventstore *eventstore.Eventstore
}

func (mig *FillFieldsForInstanceDomains) Execute(ctx context.Context, _ eventstore.Event) error {
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
		if err := projection.InstanceDomainFields.Trigger(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (mig *FillFieldsForInstanceDomains) String() string {
	return "repeatable_fill_fields_for_instance_domains"
}

func (f *FillFieldsForInstanceDomains) Check(lastRun map[string]interface{}) bool {
	return true
}
