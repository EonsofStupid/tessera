package setup

import (
	"context"
	_ "embed"

	"github.com/EonsofStupid/tessera/internal/database"
	"github.com/EonsofStupid/tessera/internal/eventstore"
)

var (
	//go:embed 20.sql
	addByUserIndexToSession string
)

type AddByUserIndexToSession struct {
	dbClient *database.DB
}

func (mig *AddByUserIndexToSession) Execute(ctx context.Context, _ eventstore.Event) error {
	_, err := mig.dbClient.ExecContext(ctx, addByUserIndexToSession)
	return err
}

func (mig *AddByUserIndexToSession) String() string {
	return "20_add_by_user_index_on_session"
}
