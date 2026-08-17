package setup

import (
	"context"
	_ "embed"

	"github.com/EonsofStupid/tessera/internal/database"
	"github.com/EonsofStupid/tessera/internal/eventstore"
)

var (
	//go:embed 63.sql
	alterResourceCounts string
)

type AlterResourceCounts struct {
	dbClient *database.DB
}

func (mig *AlterResourceCounts) Execute(ctx context.Context, _ eventstore.Event) error {
	_, err := mig.dbClient.ExecContext(ctx, alterResourceCounts)
	return err
}

func (mig *AlterResourceCounts) String() string {
	return "63_alter_resource_counts"
}
