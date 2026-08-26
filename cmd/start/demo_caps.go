package start

import (
	"context"

	nomen_domain "github.com/shippinAI/nomen/backend/v1/domain"
	"github.com/shippinAI/nomen/internal/query"
	"github.com/shippinAI/nomen/internal/zerrors"
)

type demoCapGuard struct {
	policy  nomen_domain.EditionPolicy
	queries *query.Queries
}

func (g demoCapGuard) DenyNewUser(ctx context.Context) error {
	count, err := g.queries.CountUsers(ctx, &query.UserSearchQueries{})
	if err != nil {
		return err
	}
	return demoCapError(g.policy.AllowNewUser(count))
}

func (g demoCapGuard) DenyNewOrganization(ctx context.Context) error {
	orgs, err := g.queries.SearchOrgs(ctx, &query.OrgSearchQueries{SearchRequest: query.SearchRequest{Limit: 1}}, nil)
	if err != nil {
		return err
	}
	var current uint64
	if orgs != nil {
		current = orgs.Count
	}
	return demoCapError(g.policy.AllowNewOrganization(current))
}

func (g demoCapGuard) DenyNewInstance(ctx context.Context) error {
	instances, err := g.queries.SearchInstances(ctx, &query.InstanceSearchQueries{SearchRequest: query.SearchRequest{Limit: 1}})
	if err != nil {
		return err
	}
	var current uint64
	if instances != nil {
		current = instances.Count
	}
	return demoCapError(g.policy.AllowNewInstance(current))
}

func demoCapError(denied *nomen_domain.ManagementError) error {
	if denied == nil {
		return nil
	}
	return zerrors.ThrowPermissionDenied(denied, "NOMEN-dcap", denied.Reason)
}
