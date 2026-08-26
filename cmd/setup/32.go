package setup

import (
	"context"
	_ "embed"

	"github.com/shippinAI/nomen/internal/database"
	"github.com/shippinAI/nomen/internal/eventstore"
)

var (
	//go:embed 32.sql
	addAuthSessionID string
)

type AddAuthSessionID struct {
	dbClient *database.DB
}

func (mig *AddAuthSessionID) Execute(ctx context.Context, _ eventstore.Event) error {
	_, err := mig.dbClient.ExecContext(ctx, addAuthSessionID)
	return err
}

func (mig *AddAuthSessionID) String() string {
	return "32_add_auth_sessionID"
}
