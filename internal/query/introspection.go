package query

import (
	"context"
	"database/sql"
	_ "embed"
	"sync"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/database"
	"github.com/shippinAI/nomen/internal/eventstore/handler/v2"
	"github.com/shippinAI/nomen/internal/query/projection"
	"github.com/shippinAI/nomen/internal/telemetry/tracing"
)

// introspectionTriggerHandlers slice can only be created after nomen
// is fully initialized, otherwise the handlers are nil.
// OnceValue takes care of creating the slice on the first request
// and than will always return the same slice on subsequent requests.
var introspectionTriggerHandlers = sync.OnceValue(func() []*handler.Handler {
	return append(oidcUserInfoTriggerHandlers(),
		projection.AppProjection,
		projection.OIDCSettingsProjection,
		projection.AuthNKeyProjection,
	)
})

type AppType string

const (
	AppTypeAPI  = "api"
	AppTypeOIDC = "oidc"
)

type IntrospectionClient struct {
	AppID                string
	ClientID             string
	HashedSecret         string
	AppType              AppType
	ProjectID            string
	ResourceOwner        string
	ProjectRoleAssertion bool
	PublicKeys           database.Map[[]byte]
}

//go:embed introspection_client_by_id.sql
var introspectionClientByIDQuery string

func (q *Queries) ActiveIntrospectionClientByID(ctx context.Context, clientID string, getKeys bool) (_ *IntrospectionClient, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	var (
		instanceID = authz.GetInstance(ctx).InstanceID()
		client     = new(IntrospectionClient)
	)

	err = q.client.QueryRowContext(ctx, func(row *sql.Row) error {
		return row.Scan(
			&client.AppID,
			&client.ClientID,
			&client.HashedSecret,
			&client.AppType,
			&client.ProjectID,
			&client.ResourceOwner,
			&client.ProjectRoleAssertion,
			&client.PublicKeys,
		)
	},
		introspectionClientByIDQuery,
		instanceID, clientID, getKeys,
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
