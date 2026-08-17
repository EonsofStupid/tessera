package setup

import (
	"context"

	"github.com/EonsofStupid/tessera/internal/database"
	"github.com/EonsofStupid/tessera/internal/eventstore"
	"github.com/EonsofStupid/tessera/internal/queue"
)

type RiverMigrateRepeatable struct {
	client *database.DB
}

func (mig *RiverMigrateRepeatable) Execute(ctx context.Context, _ eventstore.Event) error {
	return queue.NewMigrator(mig.client).Execute(ctx)
}

func (mig *RiverMigrateRepeatable) String() string {
	return "repeatable_migrate_river"
}

func (f *RiverMigrateRepeatable) Check(lastRun map[string]interface{}) bool {
	return true
}
