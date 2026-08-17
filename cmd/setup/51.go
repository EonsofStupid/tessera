package setup

import (
	"context"
	_ "embed"

	"github.com/EonsofStupid/tessera/internal/database"
	"github.com/EonsofStupid/tessera/internal/eventstore"
)

var (
	//go:embed 51.sql
	addRootCA string
)

type IDPTemplate6RootCA struct {
	dbClient *database.DB
}

func (mig *IDPTemplate6RootCA) Execute(ctx context.Context, _ eventstore.Event) error {
	_, err := mig.dbClient.ExecContext(ctx, addRootCA)
	return err
}

func (mig *IDPTemplate6RootCA) String() string {
	return "51_idp_templates6_add_root_ca"
}
