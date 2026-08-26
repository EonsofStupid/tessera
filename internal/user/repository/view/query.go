package view

import (
	"time"

	"github.com/shippinAI/nomen/internal/eventstore"
	"github.com/shippinAI/nomen/internal/repository/user"
	"github.com/shippinAI/nomen/internal/zerrors"
)

func UserByIDQuery(id, instanceID string, changeDate time.Time, eventTypes []eventstore.EventType) (*eventstore.SearchQueryBuilder, error) {
	if id == "" {
		return nil, zerrors.ThrowPreconditionFailed(nil, "EVENT-d8isw", "Errors.User.UserIDMissing")
	}
	return eventstore.NewSearchQueryBuilder(eventstore.ColumnsEvent).
		AwaitOpenTransactions().
		InstanceID(instanceID).
		CreationDateAfter(changeDate.Add(-1 * time.Microsecond)). // to simulate CreationDate >=
		AddQuery().
		AggregateTypes(user.AggregateType).
		AggregateIDs(id).
		EventTypes(eventTypes...).
		Builder(), nil
}
