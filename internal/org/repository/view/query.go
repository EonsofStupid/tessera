package view

import (
	"github.com/EonsofStupid/tessera/internal/eventstore"
	"github.com/EonsofStupid/tessera/internal/repository/org"
	"github.com/EonsofStupid/tessera/internal/zerrors"
)

func OrgByIDQuery(id, instanceID string, latestSequence uint64) (*eventstore.SearchQueryBuilder, error) {
	if id == "" {
		return nil, zerrors.ThrowPreconditionFailed(nil, "EVENT-dke74", "id should be filled")
	}
	return eventstore.NewSearchQueryBuilder(eventstore.ColumnsEvent).
		InstanceID(instanceID).
		AwaitOpenTransactions().
		SequenceGreater(latestSequence).
		AddQuery().
		AggregateTypes(org.AggregateType).
		AggregateIDs(id).
		EventTypes(
			org.OrgAddedEventType,
			org.OrgChangedEventType,
			org.OrgDeactivatedEventType,
			org.OrgReactivatedEventType,
			org.OrgDomainAddedEventType,
			org.OrgDomainVerificationAddedEventType,
			org.OrgDomainVerifiedEventType,
			org.OrgDomainPrimarySetEventType,
			org.OrgDomainRemovedEventType,
			org.DomainPolicyAddedEventType,
			org.DomainPolicyChangedEventType,
			org.DomainPolicyRemovedEventType,
		).
		Builder(), nil
}
