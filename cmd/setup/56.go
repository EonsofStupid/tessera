package setup

import (
	"context"
	_ "embed"

	"github.com/EonsofStupid/tessera/internal/database"
	"github.com/EonsofStupid/tessera/internal/eventstore"
)

var (
	//go:embed 56.sql
	addSAMLFederatedLogout string
)

type IDPTemplate6SAMLFederatedLogout struct {
	dbClient *database.DB
}

func (mig *IDPTemplate6SAMLFederatedLogout) Execute(ctx context.Context, _ eventstore.Event) error {
	_, err := mig.dbClient.ExecContext(ctx, addSAMLFederatedLogout)
	return err
}

func (mig *IDPTemplate6SAMLFederatedLogout) String() string {
	return "56_idp_templates6_add_saml_federated_logout"
}
