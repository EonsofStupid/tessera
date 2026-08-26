package admin

import (
	"github.com/shippinAI/nomen/internal/api/grpc/object"
	org_grpc "github.com/shippinAI/nomen/internal/api/grpc/org"
	"github.com/shippinAI/nomen/internal/query"
	"github.com/shippinAI/nomen/pkg/grpc/admin"
)

func listOrgRequestToModel(req *admin.ListOrgsRequest) (*query.OrgSearchQueries, error) {
	offset, limit, asc := object.ListQueryToModel(req.Query)
	queries, err := org_grpc.OrgQueriesToModel(req.Queries)
	if err != nil {
		return nil, err
	}
	return &query.OrgSearchQueries{
		SearchRequest: query.SearchRequest{
			Offset:        offset,
			Limit:         limit,
			SortingColumn: org_grpc.FieldNameToOrgColumn(req.SortingColumn),
			Asc:           asc,
		},
		Queries: queries,
	}, nil
}
