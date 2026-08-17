package authz

import (
	"github.com/EonsofStupid/tessera/internal/authz/repository"
	"github.com/EonsofStupid/tessera/internal/authz/repository/eventsourcing"
	"github.com/EonsofStupid/tessera/internal/crypto"
	"github.com/EonsofStupid/tessera/internal/database"
	"github.com/EonsofStupid/tessera/internal/eventstore"
	"github.com/EonsofStupid/tessera/internal/query"
)

func Start(queries *query.Queries, es *eventstore.Eventstore, dbClient *database.DB, authAlgorithm crypto.AuthAlgorithm, externalSecure bool) (repository.Repository, error) {
	return eventsourcing.Start(queries, es, dbClient, authAlgorithm, externalSecure)
}
