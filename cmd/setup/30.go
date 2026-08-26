package setup

import (
	"context"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/eventstore"
	"github.com/shippinAI/nomen/internal/query/projection"
	"github.com/shippinAI/nomen/internal/repository/instance"
)

type FillFieldsForOrgDomainVerified struct {
	eventstore *eventstore.Eventstore
}

func (mig *FillFieldsForOrgDomainVerified) Execute(ctx context.Context, _ eventstore.Event) error {
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
		if err := projection.OrgDomainVerifiedFields.Trigger(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (mig *FillFieldsForOrgDomainVerified) String() string {
	return "30_fill_fields_for_org_domain_verified"
}
