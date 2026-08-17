package setup

import (
	"context"
	_ "embed"

	"github.com/EonsofStupid/tessera/internal/database"
	"github.com/EonsofStupid/tessera/internal/eventstore"
)

var (
	//go:embed 27.sql
	addSAMLNameIDFormat string
)

type IDPTemplate6SAMLNameIDFormat struct {
	dbClient *database.DB
}

func (mig *IDPTemplate6SAMLNameIDFormat) Execute(ctx context.Context, _ eventstore.Event) error {
	_, err := mig.dbClient.ExecContext(ctx, addSAMLNameIDFormat)
	return err
}

func (mig *IDPTemplate6SAMLNameIDFormat) String() string {
	return "27_idp_templates6_add_saml_name_id_format"
}
