package authz

import (
	"github.com/shippinAI/nomen/internal/authz/repository"
	"github.com/shippinAI/nomen/internal/authz/repository/eventsourcing"
	"github.com/shippinAI/nomen/internal/crypto"
	"github.com/shippinAI/nomen/internal/database"
	"github.com/shippinAI/nomen/internal/eventstore"
	"github.com/shippinAI/nomen/internal/query"
)

func Start(queries *query.Queries, es *eventstore.Eventstore, dbClient *database.DB, authAlgorithm crypto.AuthAlgorithm, externalSecure bool) (repository.Repository, error) {
	return eventsourcing.Start(queries, es, dbClient, authAlgorithm, externalSecure)
}
