package setup

import (
	"context"

	"github.com/EonsofStupid/tessera/internal/eventstore"
	"github.com/EonsofStupid/tessera/internal/query/projection"
)

type projectionTables struct {
	es *eventstore.Eventstore

	Version string `json:"version"`
}

func (mig *projectionTables) Check(lastRun map[string]interface{}) bool {
	currentVersion, _ := lastRun["version"].(string)
	return currentVersion != mig.Version
}

func (mig *projectionTables) Execute(ctx context.Context, _ eventstore.Event) error {
	return projection.Init(ctx)
}

func (mig *projectionTables) String() string {
	return "projection_tables"
}
