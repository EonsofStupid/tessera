package setup

import (
	"context"
	_ "embed"

	"github.com/shippinAI/nomen/internal/database"
	"github.com/shippinAI/nomen/internal/eventstore"
)

var (
	//go:embed 25.sql
	addLowerFieldsToVerifiedEmail string
)

type User11AddLowerFieldsToVerifiedEmail struct {
	dbClient *database.DB
}

func (mig *User11AddLowerFieldsToVerifiedEmail) Execute(ctx context.Context, _ eventstore.Event) error {
	_, err := mig.dbClient.ExecContext(ctx, addLowerFieldsToVerifiedEmail)
	return err
}

func (mig *User11AddLowerFieldsToVerifiedEmail) String() string {
	return "25_user14_add_lower_fields_to_verified_email"
}
