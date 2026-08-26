package org

import (
	"context"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/v2/eventstore"
)

const (
	AggregateType   = "org"
	eventTypePrefix = AggregateType + "."
)

func NewAggregate(ctx context.Context, id string) *eventstore.Aggregate {
	return &eventstore.Aggregate{
		ID:       id,
		Type:     AggregateType,
		Instance: authz.GetInstance(ctx).InstanceID(),
		Owner:    authz.GetCtxData(ctx).OrgID,
	}
}
